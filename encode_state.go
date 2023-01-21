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
	"io"
	"reflect"
	"sync"
	"unicode/utf8"
)

const startDetectingCyclesAfter = 1000

var (
	// We want to keep the GC overhead as little as possible.
	// Frequent allocation and recycling of memory will cause a heavy burden to GC.
	// sync.Pool can cache objects that are not used temporarily and use them
	// directly (without reallocation) when they are needed next time.
	// This can potentially reduce the GC workload and improve performance.
	encodeStatePool sync.Pool

	// hex is the mapping from byte values to their lower-case hex digits.
	hex = "0123456789abcdef"
)

// An encodeState encodes TABLE into a bytes.Buffer.
type encodeState struct {
	// used to write the output of encoding before writing to the table buffer
	bytes.Buffer

	// t is the table we are using to organize the data encoded
	t *Table

	// scratch is a scratch buffer used by the some encoders.
	scratch [64]byte

	// Keep track of what pointers we've seen in the current recursive call
	// path, to avoid cycles that could lead to a stack overflow. Only do
	// the relatively expensive map operations if ptrLevel is larger than
	// startDetectingCyclesAfter, so that we skip the work if we're within a
	// reasonable amount of nested pointers deep.
	ptrLevel uint
	ptrSeen  map[any]struct{}
}

func newEncodeState(out io.Writer, sep string) *encodeState {
	if v := encodeStatePool.Get(); v != nil {
		es := v.(*encodeState)
		es.Reset()

		if len(es.ptrSeen) > 0 {
			panic("ptrEncoder.encode should have emptied ptrSeen via defers")
		}
		es.ptrLevel = 0
		return es
	}

	es := &encodeState{
		t:       New(out, WithSep(sep)),
		scratch: [64]byte{},
		ptrSeen: make(map[any]struct{}),
	}
	return es
}

// Flush writes the current buffer to the table and resets the buffer.
func (es *encodeState) Flush() {
	es.t.AddRow(es.String())
	es.Reset()
}

// tableError is an error wrapper type for internal use only.
// Panics with errors are wrapped in tableError so that the top-level recover
// can distinguish intentional panics from this package.
type tableError struct{ error }

func (es *encodeState) marshal(v any, opts encOpts) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(tableError); ok {
				err = e.error
			} else {
				panic(r)
			}
		}
	}()

	es.reflectValue(reflect.ValueOf(v), opts)
	return nil
}

// error aborts the encoding by panicking with err wrapped in tableError.
func (es *encodeState) error(err error) {
	panic(tableError{err})
}

func (es *encodeState) reflectValue(v reflect.Value, opts encOpts) {
	valueEncoder(v)(es, v, opts)
}

// NOTE: keep in sync with stringBytes below.
func (es *encodeState) string(s string, escapeHTML bool) {
	es.WriteByte('"')

	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if htmlSafeSet[b] || (!escapeHTML && safeSet[b]) {
				i++
				continue
			}
			if start < i {
				es.WriteString(s[start:i])
			}
			es.WriteByte('\\')
			switch b {
			case '\\', '"':
				es.WriteByte(b)
			case '\n':
				es.WriteByte('n')
			case '\r':
				es.WriteByte('r')
			case '\t':
				es.WriteByte('t')
			default:
				// This encodes bytes < 0x20 except for \t, \n and \r.
				// If escapeHTML is set, it also escapes <, >, and &
				// because they can lead to security holes when
				// user-controlled strings are rendered into TABLE
				// and served to some browsers.
				es.WriteString(`u00`)
				es.WriteByte(hex[b>>4])
				es.WriteByte(hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				es.WriteString(s[start:i])
			}
			es.WriteString(`\ufffd`)
			i += size
			start = i
			continue
		}
		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in TABLE strings,
		// but don't work in TABLE, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid TABLE to
		// escape them, so we do so unconditionally.
		// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
		if c == '\u2028' || c == '\u2029' {
			if start < i {
				es.WriteString(s[start:i])
			}
			es.WriteString(`\u202`)
			es.WriteByte(hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		es.WriteString(s[start:])
	}
	es.WriteByte('"')
}

// NOTE: keep in sync with string above.
func (es *encodeState) stringBytes(s []byte, escapeHTML bool) {
	es.WriteByte('"')

	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if htmlSafeSet[b] || (!escapeHTML && safeSet[b]) {
				i++
				continue
			}
			if start < i {
				es.Write(s[start:i])
			}
			es.WriteByte('\\')
			switch b {
			case '\\', '"':
				es.WriteByte(b)
			case '\n':
				es.WriteByte('n')
			case '\r':
				es.WriteByte('r')
			case '\t':
				es.WriteByte('t')
			default:
				// This encodes bytes < 0x20 except for \t, \n and \r.
				// If escapeHTML is set, it also escapes <, >, and &
				// because they can lead to security holes when
				// user-controlled strings are rendered into TABLE
				// and served to some browsers.
				es.WriteString(`u00`)
				es.WriteByte(hex[b>>4])
				es.WriteByte(hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRune(s[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				es.Write(s[start:i])
			}
			es.WriteString(`\ufffd`)
			i += size
			start = i
			continue
		}
		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in TABLE strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so we do so unconditionally.
		// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
		if c == '\u2028' || c == '\u2029' {
			if start < i {
				es.Write(s[start:i])
			}
			es.WriteString(`\u202`)
			es.WriteByte(hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		es.Write(s[start:])
	}
	es.WriteByte('"')
}
