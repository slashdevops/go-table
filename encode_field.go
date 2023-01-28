// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package json implements encoding and decoding of JSON as defined in
// RFC 7159. The mapping between JSON and Go values is described
// in the documentation for the Marshal and Unmarshal functions.
//
// See "JSON and Go" for an introduction to this package:
// https://golang.org/doc/articles/json_and_go.html

// NOTE: This package is a fork of the standard library's encoding/json package.
// It has been modified to support the Table data structure.
// source: https://go.dev/src/encoding/json/ or https://cs.opensource.google/go/go/+/master:src/encoding/json/
package table

import (
	"reflect"
	"sort"
	"sync"
)

func typeByIndex(t reflect.Type, index []int) reflect.Type {
	for _, i := range index {
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		t = t.Field(i).Type
	}
	return t
}

// A field represents a single field found in a struct.
type field struct {
	name      string
	nameBytes []byte // []byte(name)
	index     []int
	typ       reflect.Type
	tag       bool
	omitEmpty bool
	quoted    bool
	encoder   encoderFunc
}

// byIndex sorts field by index sequence.
type byIndex []field

func (x byIndex) Len() int { return len(x) }

func (x byIndex) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

func (x byIndex) Less(i, j int) bool {
	for k, xik := range x[i].index {
		if k >= len(x[j].index) {
			return false
		}
		if xik != x[j].index[k] {
			return xik < x[j].index[k]
		}
	}
	return len(x[i].index) < len(x[j].index)
}

// typeFields returns a list of fields that should recognize for the given type.
// The algorithm is deep-first search over the set of structs to include - the top struct
// and then any reachable anonymous structs.
func typeFields(t reflect.Type, prefix string) structFields {
	st := NewFieldStack()
	var fields []field

	if t.Kind() == reflect.Ptr {
		// if the type is a pointer, get the element type
		t = t.Elem()
	}

	// add the top struct to the stack
	st.Push(field{typ: t})

	for st.Len() > 0 {
		current := st.Pop().(field)

		if current.typ.Kind() != reflect.Struct {
			// if this not a struct, avoid to iterate over its fields
			continue
		}

		prefix := ""
		if current.name != "" {
			prefix = current.name + "."
		}

		// Scan f.typ for fields to include.
		for i := 0; i < current.typ.NumField(); i++ {
			f := current.typ.Field(i)
			ft := f.Type
			fn := f.Name

			if f.Anonymous {
				ft := f.Type

				if ft.Kind() == reflect.Pointer {
					ft = ft.Elem()
				}

				if !f.IsExported() && ft.Kind() != reflect.Struct {
					// Ignore embedded fields of unexported non-struct types.
					continue
				}
				// Do not ignore embedded fields of unexported struct types
				// since they may have exported fields.
			} else if !f.IsExported() {
				// Ignore unexported non-embedded fields.
				continue
			}

			tag := f.Tag.Get(structTagName)
			if tag == "-" {
				continue
			}

			name, opts := parseTag(tag)
			if !isValidTag(name) {
				name = ""
			}

			if ft.Name() == "" && ft.Kind() == reflect.Pointer {
				// Follow pointer.
				ft = ft.Elem()
			}

			tagged := name != ""
			if name == "" {
				name = fn
			}

			// Only strings, floats, integers, and booleans can be quoted.
			quoted := false
			if opts.Contains("string") {
				switch ft.Kind() {
				case reflect.Bool,
					reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
					reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
					reflect.Float32, reflect.Float64,
					reflect.String:
					quoted = true
				}
			}

			fname := prefix + name

			// this fills the index array taking into account the nested structs
			// each level is added to the index array
			index := make([]int, len(current.index)+1)
			copy(index, current.index)
			index[len(current.index)] = i

			// we don't want to add fields of struct type
			// as a field of the table, just the fields of the struct
			k := ft.Kind()
			if name != "" && !f.Anonymous && k != reflect.Struct {
				field := field{
					name:      fname,
					tag:       tagged,
					typ:       ft,
					index:     index,
					omitEmpty: opts.Contains("omitempty"),
					quoted:    quoted,
					encoder:   typeEncoder(ft),
				}
				field.nameBytes = []byte(field.name)

				fields = append(fields, field)
			}

			if k == reflect.Struct {
				typeFields(ft, fn+".")
			}

			// push analyzed field type to the stack
			st.Push(field{name: fname, typ: ft, index: index})
		}
	}

	sort.Slice(fields, func(i, j int) bool {
		x := fields
		// sort field by name, breaking ties with depth, then
		// breaking ties with "name came from json tag", then
		// breaking ties with index sequence.
		if x[i].name != x[j].name {
			return x[i].name < x[j].name
		}
		if len(x[i].index) != len(x[j].index) {
			return len(x[i].index) < len(x[j].index)
		}
		if x[i].tag != x[j].tag {
			return x[i].tag
		}
		return byIndex(x).Less(i, j)
	})

	// The fields are sorted in primary order of name, secondary order
	// of field index length. Loop over names; for each name, delete
	// hidden fields by choosing the one dominant field that survives.
	out := fields[:0]
	for advance, i := 0, 0; i < len(fields); i += advance {
		// One iteration per name.
		// Find the sequence of fields with the name of this first field.
		fi := fields[i]
		name := fi.name
		for advance = 1; i+advance < len(fields); advance++ {
			fj := fields[i+advance]
			if fj.name != name {
				break
			}
		}
		if advance == 1 { // Only one field with this name
			out = append(out, fi)
			continue
		}
		dominant, ok := dominantField(fields[i : i+advance])
		if ok {
			out = append(out, dominant)
		}
	}

	fields = out
	sort.Sort(byIndex(fields))

	for i := range fields {
		f := &fields[i]
		f.encoder = typeEncoder(typeByIndex(t, f.index))
	}

	nameIndex := make(map[string]int, len(fields))
	for i, field := range fields {
		nameIndex[field.name] = i
	}
	return structFields{fields, nameIndex}
}

// dominantField looks through the fields, all of which are known to
// have the same name, to find the single field that dominates the
// others using Go's embedding rules, modified by the presence of
// TABLE tags. If there are multiple top-level fields, the boolean
// will be false: This condition is an error in Go and we skip all
// the fields.
func dominantField(fields []field) (field, bool) {
	// The fields are sorted in increasing index-length order, then by presence of tag.
	// That means that the first field is the dominant one. We need only check
	// for error cases: two fields at top level, either both tagged or neither tagged.
	if len(fields) > 1 && len(fields[0].index) == len(fields[1].index) && fields[0].tag == fields[1].tag {
		return field{}, false
	}
	return fields[0], true
}

var fieldCache sync.Map // map[reflect.Type]structFields

// cachedTypeFields is like typeFields but uses a cache to avoid repeated work.
func cachedTypeFields(t reflect.Type) structFields {
	if f, ok := fieldCache.Load(t); ok {
		return f.(structFields)
	}
	f, _ := fieldCache.LoadOrStore(t, typeFields(t, ""))
	return f.(structFields)
}
