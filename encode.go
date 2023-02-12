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
	"bytes"
	"encoding"
	"reflect"
	"sync"
)

// Marshaler is the interface implemented by types that
// can marshal themselves into valid Table.
type Marshaler interface {
	MarshalTable() ([]byte, error)
}

var (
	marshalerType     = reflect.TypeOf((*Marshaler)(nil)).Elem()
	textMarshalerType = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
)

// Marshal returns the table encoding of v.
// sep is optional and defaults to "\t".  This is the separator used
// between fields.  If sep is empty, the default is used.
func Marshal(v any, opts ...MarshalOption) ([]byte, error) {
	o := marshalOptions{
		sep:          "\t",
		flattenArray: false,
		flattenMap:   false,
	}

	for _, opt := range opts {
		opt(&o)
	}

	buf := new(bytes.Buffer)
	es := newEncodeState(buf, o.sep)

	err := es.marshal(v, encOpts{sep: o.sep, flattenArray: o.flattenArray, flattenMap: o.flattenMap, escapeHTML: true})
	if err != nil {
		return nil, err
	}

	es.Flush()
	es.t.Render()

	out := append([]byte(nil), buf.Bytes()...)

	return out, nil
}

type encOpts struct {
	// sep is the separator between fields.
	sep string

	// flatten arrays and slices
	flattenArray bool

	// prefix
	prefix string

	// flatten maps
	flattenMap bool

	// quoted causes primitive fields to be encoded inside Table strings.
	quoted bool

	// escapeHTML causes '<', '>', and '&' to be escaped in Table strings.
	escapeHTML bool
}

// encoderFunc is the return type of newTypeEncoder and all the typeEncoder
type encoderFunc func(e *encodeState, v reflect.Value, opts encOpts)

// encoderCache is a cache of encoderFuncs for reflect.Types.
var encoderCache sync.Map // map[reflect.Type]encoderFunc

func valueEncoder(v reflect.Value) encoderFunc {
	if !v.IsValid() {
		return invalidValueEncoder
	}
	return typeEncoder(v.Type())
}

func typeEncoder(t reflect.Type) encoderFunc {
	if fi, ok := encoderCache.Load(t); ok {
		return fi.(encoderFunc)
	}

	// To deal with recursive types, populate the map with an
	// indirect func before we build it. This type waits on the
	// real func (f) to be ready and then calls it. This indirect
	// func is only used for recursive types.
	var (
		wg sync.WaitGroup
		f  encoderFunc
	)
	wg.Add(1)
	fi, loaded := encoderCache.LoadOrStore(t, encoderFunc(func(e *encodeState, v reflect.Value, opts encOpts) {
		wg.Wait()
		f(e, v, opts)
	}))
	if loaded {
		return fi.(encoderFunc)
	}

	// Compute the real encoder and replace the indirect func with it.
	f = newTypeEncoder(t, true)
	wg.Done()

	encoderCache.Store(t, f)
	return f
}

// newTypeEncoder constructs an encoderFunc for a type.
// The returned encoder only checks CanAddr when allowAddr is true.
func newTypeEncoder(t reflect.Type, allowAddr bool) encoderFunc {
	// If we have a non-pointer value whose type implements
	// Marshaler with a value receiver, then we're better off taking
	// the address of the value - otherwise we end up with an
	// allocation as we cast the value to an interface.
	if t.Kind() != reflect.Pointer && allowAddr && reflect.PointerTo(t).Implements(marshalerType) {
		return newCondAddrEncoder(addrMarshalerEncoder, newTypeEncoder(t, false))
	}
	if t.Implements(marshalerType) {
		return marshalerEncoder
	}
	if t.Kind() != reflect.Pointer && allowAddr && reflect.PointerTo(t).Implements(textMarshalerType) {
		return newCondAddrEncoder(addrTextMarshalerEncoder, newTypeEncoder(t, false))
	}
	if t.Implements(textMarshalerType) {
		return textMarshalerEncoder
	}

	switch t.Kind() {
	case reflect.Bool:
		return boolEncoder
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intEncoder
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uintEncoder
	case reflect.Float32:
		return float32Encoder
	case reflect.Float64:
		return float64Encoder
	case reflect.String:
		return stringEncoder
	case reflect.Interface:
		return interfaceEncoder
	case reflect.Struct:
		return newStructEncoder(t)
	case reflect.Map:
		return newMapEncoder(t)
	case reflect.Slice:
		return newSliceEncoder(t)
	case reflect.Array:
		return newArrayEncoder(t)
	case reflect.Pointer:
		return newPtrEncoder(t)
	default:
		return unsupportedTypeEncoder
	}
}
