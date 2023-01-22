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
	"encoding/base64"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
)

func invalidValueEncoder(es *encodeState, v reflect.Value, _ encOpts) {
	es.WriteString("null")
}

// isValidNumber reports whether s is a valid JSON number literal.
func isValidNumber(s string) bool {
	// This function implements the JSON numbers grammar.
	// See https://tools.ietf.org/html/rfc7159#section-6
	// and https://www.json.org/img/number.png

	if s == "" {
		return false
	}

	// Optional -
	if s[0] == '-' {
		s = s[1:]
		if s == "" {
			return false
		}
	}

	// Digits
	switch {
	default:
		return false

	case s[0] == '0':
		s = s[1:]

	case '1' <= s[0] && s[0] <= '9':
		s = s[1:]
		for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
			s = s[1:]
		}
	}

	// . followed by 1 or more digits.
	if len(s) >= 2 && s[0] == '.' && '0' <= s[1] && s[1] <= '9' {
		s = s[2:]
		for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
			s = s[1:]
		}
	}

	// e or E followed by an optional - or + and
	// 1 or more digits.
	if len(s) >= 2 && (s[0] == 'e' || s[0] == 'E') {
		s = s[1:]
		if s[0] == '+' || s[0] == '-' {
			s = s[1:]
			if s == "" {
				return false
			}
		}
		for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
			s = s[1:]
		}
	}

	// Make sure we are at the end.
	return s == ""
}

type condAddrEncoder struct {
	canAddrEnc, elseEnc encoderFunc
}

func (ce condAddrEncoder) encode(e *encodeState, v reflect.Value, opts encOpts) {
	if v.CanAddr() {
		ce.canAddrEnc(e, v, opts)
	} else {
		ce.elseEnc(e, v, opts)
	}
}

// newCondAddrEncoder returns an encoder that checks whether its value
// CanAddr and delegates to canAddrEnc if so, else to elseEnc.
func newCondAddrEncoder(canAddrEnc, elseEnc encoderFunc) encoderFunc {
	enc := condAddrEncoder{canAddrEnc: canAddrEnc, elseEnc: elseEnc}
	return enc.encode
}

func stringEncoder(es *encodeState, v reflect.Value, opts encOpts) {
	if v.Type() == numberType {
		numStr := v.String()
		// In Go1.5 the empty string encodes to "0", while this is not a valid number literal
		// we keep compatibility so check validity after this.
		if numStr == "" {
			numStr = "0" // Number's zero-val
		}
		if !isValidNumber(numStr) {
			es.error(fmt.Errorf("json: invalid number literal %q", numStr))
		}
		if opts.quoted {
			es.WriteByte('"')
		}
		es.WriteString(numStr)
		if opts.quoted {
			es.WriteByte('"')
		}
		return
	}
	if opts.quoted {
		buf := new(bytes.Buffer)
		e2 := newEncodeState(buf, es.t.sep)

		// Since we encode the string twice, we only need to escape HTML
		// the first time.
		e2.string(v.String(), opts.escapeHTML)
		es.stringBytes(e2.Bytes(), false)
	} else {
		es.string(v.String(), opts.escapeHTML)
	}
}

// A Number represents a JSON number literal.
type Number string

// String returns the literal text of the number.
func (n Number) String() string { return string(n) }

// Float64 returns the number as a float64.
func (n Number) Float64() (float64, error) {
	return strconv.ParseFloat(string(n), 64)
}

// Int64 returns the number as an int64.
func (n Number) Int64() (int64, error) {
	return strconv.ParseInt(string(n), 10, 64)
}

var numberType = reflect.TypeOf(Number(""))

// An UnsupportedTypeError is returned by Marshal when attempting
// to encode an unsupported value type.
type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "table: unsupported type: " + e.Type.String()
}

// An UnsupportedValueError is returned by Marshal when attempting
// to encode an unsupported value.
type UnsupportedValueError struct {
	Value reflect.Value
	Str   string
}

func (e *UnsupportedValueError) Error() string {
	return "table: unsupported value: " + e.Str
}

// Before Go 1.2, an InvalidUTF8Error was returned by Marshal when
// attempting to encode a string value with invalid UTF-8 sequences.
// As of Go 1.2, Marshal instead coerces the string to valid UTF-8 by
// replacing invalid bytes with the Unicode replacement rune U+FFFD.
//
// Deprecated: No longer used; kept for compatibility.
type InvalidUTF8Error struct {
	S string // the whole string value that caused the error
}

func (e *InvalidUTF8Error) Error() string {
	return "table: invalid UTF-8 in string: " + strconv.Quote(e.S)
}

// A MarshalerError represents an error from calling a MarshalJSON or MarshalText method.
type MarshalerError struct {
	Type       reflect.Type
	Err        error
	sourceFunc string
}

func (e *MarshalerError) Error() string {
	srcFunc := e.sourceFunc
	if srcFunc == "" {
		srcFunc = "MarshalJSON"
	}
	return "table: error calling " + srcFunc +
		" for type " + e.Type.String() +
		": " + e.Err.Error()
}

// Unwrap returns the underlying error.
func (e *MarshalerError) Unwrap() error { return e.Err }

type floatEncoder int // number of bits

func (bits floatEncoder) encode(es *encodeState, v reflect.Value, opts encOpts) {
	f := v.Float()
	if math.IsInf(f, 0) || math.IsNaN(f) {
		es.error(&UnsupportedValueError{v, strconv.FormatFloat(f, 'g', -1, int(bits))})
	}

	// Convert as if by ES6 number to string conversion.
	// This matches most other JSON generators.
	// See golang.org/issue/6384 and golang.org/issue/14135.
	// Like fmt %g, but the exponent cutoffs are different
	// and exponents themselves are not padded to two digits.
	b := es.scratch[:0]
	abs := math.Abs(f)
	fmt := byte('f')
	// Note: Must use float32 comparisons for underlying float32 value to get precise cutoffs right.
	if abs != 0 {
		if bits == 64 && (abs < 1e-6 || abs >= 1e21) || bits == 32 && (float32(abs) < 1e-6 || float32(abs) >= 1e21) {
			fmt = 'e'
		}
	}
	b = strconv.AppendFloat(b, f, fmt, -1, int(bits))
	if fmt == 'e' {
		// clean up e-09 to e-9
		n := len(b)
		if n >= 4 && b[n-4] == 'e' && b[n-3] == '-' && b[n-2] == '0' {
			b[n-2] = b[n-1]
			b = b[:n-1]
		}
	}

	if opts.quoted {
		es.WriteByte('"')
	}

	es.Write(b)

	if opts.quoted {
		es.WriteByte('"')
	}
}

var (
	float32Encoder = (floatEncoder(32)).encode
	float64Encoder = (floatEncoder(64)).encode
)

func marshalerEncoder(es *encodeState, v reflect.Value, opts encOpts) {
	if v.Kind() == reflect.Pointer && v.IsNil() {
		es.WriteString("null")
		return
	}
	m, ok := v.Interface().(Marshaler)
	if !ok {
		es.WriteString("null")
		return
	}
	b, err := m.MarshalTable()
	_ = b
	// if err == nil {
	// 	// copy JSON into buffer, checking validity.
	// 	err = compact(&e.Buffer, b, opts.escapeHTML)
	// }
	if err != nil {
		es.error(&MarshalerError{v.Type(), err, "MarshalJSON"})
	}
}

func addrMarshalerEncoder(es *encodeState, v reflect.Value, opts encOpts) {
	va := v.Addr()
	if va.IsNil() {
		es.WriteString("null")
		return
	}
	m := va.Interface().(Marshaler)
	b, err := m.MarshalTable()
	_ = b
	// if err == nil {
	// 	err = compact(&e.Buffer, b, opts.escapeHTML)
	// }
	if err != nil {
		es.error(&MarshalerError{v.Type(), err, "MarshalJSON"})
	}
}

func textMarshalerEncoder(es *encodeState, v reflect.Value, opts encOpts) {
	if v.Kind() == reflect.Pointer && v.IsNil() {
		es.WriteString("null")
		return
	}
	m, ok := v.Interface().(encoding.TextMarshaler)
	if !ok {
		es.WriteString("null")
		return
	}
	b, err := m.MarshalText()
	if err != nil {
		es.error(&MarshalerError{v.Type(), err, "MarshalText"})
	}
	es.stringBytes(b, opts.escapeHTML)
}

func addrTextMarshalerEncoder(es *encodeState, v reflect.Value, opts encOpts) {
	va := v.Addr()
	if va.IsNil() {
		es.WriteString("null")
		return
	}
	m := va.Interface().(encoding.TextMarshaler)
	b, err := m.MarshalText()
	if err != nil {
		es.error(&MarshalerError{v.Type(), err, "MarshalText"})
	}
	es.stringBytes(b, opts.escapeHTML)
}

func boolEncoder(es *encodeState, v reflect.Value, opts encOpts) {
	if opts.quoted {
		es.WriteByte('"')
	}

	if v.Bool() {
		es.WriteString("true")
	} else {
		es.WriteString("false")
	}

	if opts.quoted {
		es.WriteByte('"')
	}
}

func intEncoder(es *encodeState, v reflect.Value, opts encOpts) {
	b := strconv.AppendInt(es.scratch[:0], v.Int(), 10)

	if opts.quoted {
		es.WriteByte('"')
	}

	es.Write(b)

	if opts.quoted {
		es.WriteByte('"')
	}
}

func uintEncoder(es *encodeState, v reflect.Value, opts encOpts) {
	b := strconv.AppendUint(es.scratch[:0], v.Uint(), 10)

	if opts.quoted {
		es.WriteByte('"')
	}

	es.Write(b)

	if opts.quoted {
		es.WriteByte('"')
	}
}

func interfaceEncoder(es *encodeState, v reflect.Value, opts encOpts) {
	if v.IsNil() {
		es.WriteString("null")
		return
	}
	es.reflectValue(v.Elem(), opts)
}

func unsupportedTypeEncoder(e *encodeState, v reflect.Value, _ encOpts) {
	e.error(&UnsupportedTypeError{v.Type()})
}

type structEncoder struct {
	fields structFields
}

type structFields struct {
	list      []field
	nameIndex map[string]int
}

func (se structEncoder) encode(es *encodeState, v reflect.Value, opts encOpts) {
	headers := make([]string, 0, len(se.fields.list))
	for i := 0; i < len(se.fields.list); i++ {
		f := se.fields.list[i]
		if f.omitEmpty && isEmptyValue(v.FieldByIndex(f.index)) {
			continue
		}
		headers = append(headers, f.name)
		es.reflectValue(v.FieldByIndex(f.index), opts)
		if i < len(se.fields.list)-1 {
			es.WriteString(opts.sep)
		}
	}
	es.t.SetHeader(headers)
}

func newStructEncoder(t reflect.Type) encoderFunc {
	se := structEncoder{fields: cachedTypeFields(t)}
	return se.encode
}

type mapEncoder struct {
	elemEnc encoderFunc
}

func (me mapEncoder) encode(es *encodeState, v reflect.Value, opts encOpts) {
	if v.IsNil() {
		es.WriteString("null")
		return
	}

	if es.ptrLevel++; es.ptrLevel > startDetectingCyclesAfter {
		// We're a large number of nested ptrEncoder.encode calls deep;
		// start checking if we've run into a pointer cycle.
		ptr := v.UnsafePointer()
		if _, ok := es.ptrSeen[ptr]; ok {
			es.error(&UnsupportedValueError{v, fmt.Sprintf("encountered a cycle via %s", v.Type())})
		}
		es.ptrSeen[ptr] = struct{}{}
		defer delete(es.ptrSeen, ptr)
	}

	es.WriteByte('\'')
	es.WriteByte('{')

	// Extract and sort the keys.
	sv := make([]reflectWithString, v.Len())
	mi := v.MapRange()
	for i := 0; mi.Next(); i++ {
		sv[i].k = mi.Key()
		sv[i].v = mi.Value()
		if err := sv[i].resolve(); err != nil {
			es.error(fmt.Errorf("table: encoding error for type %q: %q", v.Type().String(), err.Error()))
		}
	}
	sort.Slice(sv, func(i, j int) bool { return sv[i].ks < sv[j].ks })

	for i, kv := range sv {
		if i > 0 {
			es.WriteByte(',')
		}
		es.string(kv.ks, opts.escapeHTML)
		es.WriteByte(':')
		me.elemEnc(es, kv.v, opts)
	}
	es.WriteByte('}')
	es.WriteByte('\'')

	es.ptrLevel--
}

func newMapEncoder(t reflect.Type) encoderFunc {
	switch t.Key().Kind() {
	case reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
	default:
		if !t.Key().Implements(textMarshalerType) {
			return unsupportedTypeEncoder
		}
	}
	me := mapEncoder{typeEncoder(t.Elem())}
	return me.encode
}

func encodeByteSlice(es *encodeState, v reflect.Value, _ encOpts) {
	if v.IsNil() {
		es.WriteString("null")
		return
	}
	s := v.Bytes()

	es.WriteByte('"')
	encodedLen := base64.StdEncoding.EncodedLen(len(s))
	if encodedLen <= len(es.scratch) {
		// If the encoded bytes fit in e.scratch, avoid an extra
		// allocation and use the cheaper Encoding.Encode.
		dst := es.scratch[:encodedLen]
		base64.StdEncoding.Encode(dst, s)
		es.Write(dst)
	} else if encodedLen <= 1024 {
		// The encoded bytes are short enough to allocate for, and
		// Encoding.Encode is still cheaper.
		dst := make([]byte, encodedLen)
		base64.StdEncoding.Encode(dst, s)
		es.Write(dst)
	} else {
		// The encoded bytes are too long to cheaply allocate, and
		// Encoding.Encode is no longer noticeably cheaper.
		enc := base64.NewEncoder(base64.StdEncoding, es)
		enc.Write(s)
		enc.Close()
	}
	es.WriteByte('"')
}

// sliceEncoder just wraps an arrayEncoder, checking to make sure the value isn't nil.
type sliceEncoder struct {
	arrayEnc encoderFunc
}

func (se sliceEncoder) encode(es *encodeState, v reflect.Value, opts encOpts) {
	if v.IsNil() {
		es.WriteString("null")
		return
	}
	if es.ptrLevel++; es.ptrLevel > startDetectingCyclesAfter {
		// We're a large number of nested ptrEncoder.encode calls deep;
		// start checking if we've run into a pointer cycle.
		// Here we use a struct to memorize the pointer to the first element of the slice
		// and its length.
		ptr := struct {
			ptr interface{} // always an unsafe.Pointer, but avoids a dependency on package unsafe
			len int
		}{v.UnsafePointer(), v.Len()}
		if _, ok := es.ptrSeen[ptr]; ok {
			es.error(&UnsupportedValueError{v, fmt.Sprintf("encountered a cycle via %s", v.Type())})
		}
		es.ptrSeen[ptr] = struct{}{}
		defer delete(es.ptrSeen, ptr)
	}
	se.arrayEnc(es, v, opts)
	es.ptrLevel--
}

func newSliceEncoder(t reflect.Type) encoderFunc {
	// Byte slices get special treatment; arrays don't.
	if t.Elem().Kind() == reflect.Uint8 {
		p := reflect.PointerTo(t.Elem())
		if !p.Implements(marshalerType) && !p.Implements(textMarshalerType) {
			return encodeByteSlice
		}
	}
	enc := sliceEncoder{newArrayEncoder(t)}
	return enc.encode
}

type arrayEncoder struct {
	elemEnc encoderFunc
}

func (ae arrayEncoder) encode(es *encodeState, v reflect.Value, opts encOpts) {
	es.WriteByte('\'')
	es.WriteByte('[')
	n := v.Len()
	for i := 0; i < n; i++ {
		if i > 0 {
			es.WriteByte(',')
		}
		ae.elemEnc(es, v.Index(i), opts)
	}
	es.WriteByte(']')
	es.WriteByte('\'')
}

func newArrayEncoder(t reflect.Type) encoderFunc {
	enc := arrayEncoder{typeEncoder(t.Elem())}
	return enc.encode
}

type ptrEncoder struct {
	elemEnc encoderFunc
}

func (pe ptrEncoder) encode(es *encodeState, v reflect.Value, opts encOpts) {
	if v.IsNil() {
		es.WriteString("null")
		return
	}
	if es.ptrLevel++; es.ptrLevel > startDetectingCyclesAfter {
		// We're a large number of nested ptrEncoder.encode calls deep;
		// start checking if we've run into a pointer cycle.
		ptr := v.Interface()
		if _, ok := es.ptrSeen[ptr]; ok {
			es.error(&UnsupportedValueError{v, fmt.Sprintf("encountered a cycle via %s", v.Type())})
		}
		es.ptrSeen[ptr] = struct{}{}
		defer delete(es.ptrSeen, ptr)
	}
	pe.elemEnc(es, v.Elem(), opts)
	es.ptrLevel--
}

func newPtrEncoder(t reflect.Type) encoderFunc {
	enc := ptrEncoder{typeEncoder(t.Elem())}
	return enc.encode
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	}
	return false
}

type reflectWithString struct {
	k  reflect.Value
	v  reflect.Value
	ks string
}

func (w *reflectWithString) resolve() error {
	if w.k.Kind() == reflect.String {
		w.ks = w.k.String()
		return nil
	}
	if tm, ok := w.k.Interface().(encoding.TextMarshaler); ok {
		if w.k.Kind() == reflect.Pointer && w.k.IsNil() {
			return nil
		}
		buf, err := tm.MarshalText()
		w.ks = string(buf)
		return err
	}
	switch w.k.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		w.ks = strconv.FormatInt(w.k.Int(), 10)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		w.ks = strconv.FormatUint(w.k.Uint(), 10)
		return nil
	}
	panic("unexpected map key type")
}
