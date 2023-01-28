package table

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func Test_typeFields(t *testing.T) {
	type Example struct {
		Name  string `table:"name,omitempty"`
		Count int    `table:"count,omitempty"`
	}

	given := Example{
		Name:  "John",
		Count: 1,
	}

	want := structFields{
		list: []field{
			{
				name:      "name",
				nameBytes: []byte("name"),
				index:     []int{0},
				tag:       true,
				typ:       reflect.TypeOf("string"),
				omitEmpty: true,
				quoted:    false,
				encoder:   typeEncoder(reflect.TypeOf("string")),
			},
			{
				name:      "count",
				nameBytes: []byte("count"),
				index:     []int{1},
				tag:       true,
				typ:       reflect.TypeOf(int(1)),
				omitEmpty: true,
				quoted:    false,
				encoder:   typeEncoder(reflect.TypeOf(int(1))),
			},
		},
		nameIndex: map[string]int{"name": 0, "count": 1},
	}

	t.Run("no pointers", func(t *testing.T) {
		got := typeFields(reflect.TypeOf(given), "")

		opts := []cmp.Option{
			cmp.AllowUnexported(structFields{}),
			cmp.AllowUnexported(field{}),
			cmpopts.IgnoreFields(field{}, "typ"),
			cmpopts.IgnoreFields(field{}, "encoder"),
		}

		if diff := cmp.Diff(want, got, opts...); diff != "" {
			t.Errorf("typeFields() mismatch (-want +got):\n%s", diff)
		}
	})
}

func Test_typeFieldTwo(t *testing.T) {
	type Link struct {
		HRef string `table:"href,omitempty"`
		Name string `table:"name,omitempty"`
	}

	type Object struct {
		Self  *Link `table:"self,omitempty"`
		Hooks *Link `table:"hooks,omitempty"`
	}

	type Repository struct {
		Links *Object `table:"links,omitempty"`
	}

	r := Repository{
		Links: &Object{
			Self: &Link{
				HRef: "https://api.bitbucket.org/2.0/repositories/username/repo",
			},
			Hooks: &Link{
				HRef: "https://api.bitbucket.org/2.0/repositories/username/repo/hooks",
				Name: "hooks",
			},
		},
	}

	fields := []field{
		{
			name:      "links.self.href",
			nameBytes: []byte("links.self.href"),
			index:     []int{0, 0, 0},
			tag:       true,
			typ:       reflect.TypeOf("string"),
			omitEmpty: true,
			quoted:    false,
			encoder:   typeEncoder(reflect.TypeOf("string")),
		},
		{
			name:      "links.self.name",
			nameBytes: []byte("links.self.name"),
			index:     []int{0, 0, 1},
			tag:       true,
			typ:       reflect.TypeOf("string"),
			omitEmpty: true,
			quoted:    false,
			encoder:   typeEncoder(reflect.TypeOf("string")),
		},
		{
			name:      "links.hooks.href",
			nameBytes: []byte("links.hooks.href"),
			index:     []int{0, 1, 0},
			tag:       true,
			typ:       reflect.TypeOf("string"),
			omitEmpty: true,
			quoted:    false,
			encoder:   typeEncoder(reflect.TypeOf("string")),
		},
		{
			name:      "links.hooks.name",
			nameBytes: []byte("links.hooks.name"),
			index:     []int{0, 1, 1},
			tag:       true,
			typ:       reflect.TypeOf("string"),
			omitEmpty: true,
			quoted:    false,
			encoder:   typeEncoder(reflect.TypeOf("string")),
		},
	}

	want := structFields{
		list: fields,
		nameIndex: map[string]int{
			"links.self.href": 0, "links.self.name": 1, "links.hooks.href": 2, "links.hooks.name": 3,
		},
	}

	got := typeFields(reflect.ValueOf(r).Type(), "")

	opts := []cmp.Option{
		cmp.AllowUnexported(structFields{}),
		cmp.AllowUnexported(field{}),
		cmpopts.IgnoreFields(field{}, "typ"),
		cmpopts.IgnoreFields(field{}, "encoder"),
	}

	if diff := cmp.Diff(want, got, opts...); diff != "" {
		t.Errorf("typeFields() mismatch (-want +got):\n%s", diff)
	}
}
