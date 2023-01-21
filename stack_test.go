package table

import (
	"reflect"
	"testing"
)

func TestStack(t *testing.T) {
	s := NewStack()

	want := &stack{
		v: make([]*field, 0),
	}

	if !reflect.DeepEqual(s, want) {
		t.Errorf("Expected stack to be %v, got %v", want, s)
	}

	if s.Len() != 0 {
		t.Errorf("Expected stack length to be 0, got %d", s.Len())
	}

	s.Push(field{
		name:      "Field1",
		nameBytes: []byte("Field1"),
		index:     []int{0},
		tag:       true,
		typ:       reflect.TypeOf("string"),
		omitEmpty: false,
		quoted:    false,
		encoder:   typeEncoder(reflect.TypeOf("string")),
	})

	s.Push(field{
		name:      "Field2",
		nameBytes: []byte("Field2"),
		index:     []int{1},
		tag:       true,
		typ:       reflect.TypeOf(int(1)),
		omitEmpty: false,
		quoted:    false,
		encoder:   typeEncoder(reflect.TypeOf(int(1))),
	})

	if s.Len() != 2 {
		t.Errorf("Expected stack length to be 2, got %d", s.Len())
	}

	e2 := s.Pop()
	if s.Len() != 1 {
		t.Errorf("Expected stack length to be 1, got %d", s.Len())
	}
	if e2.(field).name != "Field2" {
		t.Errorf("Expected element name to be 'Field2', got '%s'", e2.(field).name)
	}

	e1 := s.Pop()
	if s.Len() != 0 {
		t.Errorf("Expected stack length to be 0, got %d", s.Len())
	}
	if e1.(field).name != "Field1" {
		t.Errorf("Expected element name to be 'Field1', got '%s'", e1.(field).name)
	}
}
