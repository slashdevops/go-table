package table

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestMarshal(t *testing.T) {
	var sp string = "pointer to string"

	type st struct {
		A int               `table:"a"`
		B string            `table:"b"`
		C *string           `table:"c"`
		D []string          `table:"d"`
		E map[string]string `table:"e"`
	}

	type complex struct {
		Field1 string  `table:"field1"`
		Field2 int     `table:"-"`
		Field3 float64 `table:"field3"`
		field4 string
		Field5 *string
		Field6 *st `table:"field6"`
	}

	cmp := complex{
		Field1: "field1",
		Field2: 2,
		Field3: float64(3.3),
		field4: "field4",
		Field5: &sp,
		Field6: &st{
			A: 1,
			B: "b",
			C: &sp,
			D: []string{"d1", "d2"},
			E: map[string]string{"e1": "e1", "e2": "e2"},
		},
	}

	tests := []struct {
		v    any
		opts MarshalOption
		want []byte
	}{
		{v: nil, opts: WithFieldSeparator(","), want: []byte("Unknown1\nnull\n")},
		{v: true, opts: WithFieldSeparator(","), want: []byte("Unknown1\ntrue\n")},
		{v: false, opts: WithFieldSeparator(","), want: []byte("Unknown1\nfalse\n")},
		{v: 12, opts: WithFieldSeparator(","), want: []byte("Unknown1\n12\n")},
		{v: uint64(111), opts: WithFieldSeparator(","), want: []byte("Unknown1\n111\n")},
		{v: float32(32.32), opts: WithFieldSeparator(","), want: []byte("Unknown1\n32.32\n")},
		{v: float64(64.64), opts: WithFieldSeparator(","), want: []byte("Unknown1\n64.64\n")},
		{v: "this is a string", opts: WithFieldSeparator(","), want: []byte("Unknown1\n\"this is a string\"\n")},
		{v: [5]int{1, 2, 3, 4, 5}, opts: WithFieldSeparator(","), want: []byte("Unknown1\n'[1,2,3,4,5]'\n")},
		{v: []int{1, 2, 3, 4, 5}, opts: WithFieldSeparator(","), want: []byte("Unknown1\n'[1,2,3,4,5]'\n")},
		{v: []float64{1.5, 2.5, 3.5, 4.5, 5.5}, opts: WithFieldSeparator(","), want: []byte("Unknown1\n'[1.5,2.5,3.5,4.5,5.5]'\n")},
		{v: []string{"this is a string", "b", "c"}, opts: WithFieldSeparator(","), want: []byte("Unknown1\n'[\"this is a string\",\"b\",\"c\"]'\n")},
		{v: map[string]int{"foo": 1, "bar": 2}, opts: WithFieldSeparator(","), want: []byte("Unknown1\n'{\"bar\":2,\"foo\":1}'\n")},
		{v: map[int]string{1: "foo", 2: "bar"}, opts: WithFieldSeparator(","), want: []byte("Unknown1\n'{\"1\":\"foo\",\"2\":\"bar\"}'\n")},
		{v: &[5]int{1, 2, 3, 4, 5}, opts: WithFieldSeparator(","), want: []byte("Unknown1\n'[1,2,3,4,5]'\n")},
		{
			v:    &st{A: 1, B: "this is a string", C: &sp, D: []string{"one", "two"}, E: map[string]string{"foo": "one", "bar": "two"}},
			opts: WithFieldSeparator(","),
			want: []byte("a,b,c,d,e\n1,\"this is a string\",\"pointer to string\",'[\"one\",\"two\"]','{\"bar\":\"two\",\"foo\":\"one\"}'\n"),
		},
		{
			v:    cmp,
			opts: WithFieldSeparator(","),
			want: []byte("field1,field3,Field5,field6.a,field6.b,field6.c,field6.d,field6.e\n\"field1\",3.3,\"pointer to string\",1,\"b\",\"pointer to string\",'[\"d1\",\"d2\"]','{\"e1\":\"e1\",\"e2\":\"e2\"}'\n"),
		},
		// {
		// 	v:    cmp,
		// 	opts: WithFieldSeparator(","),
		// 	want: []byte("field1,float,Field5,field6.a,field6.b,field6.c,field6.d,field6.e\n\"field1\",3.3,\"pointer to string\",1,\"b\",\"pointer to string\",'[\"d1\",\"d2\"]','{\"e1\":\"e1\",\"e2\":\"e2\"}'\n"),
		// },
	}

	for _, tt := range tests {
		var name string

		if tt.v == nil {
			name = "nil"
		} else {
			name = reflect.TypeOf(tt.v).String()
		}

		t.Run(name, func(t *testing.T) {
			got, err := Marshal(tt.v, tt.opts)
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Marshal() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func TestEmptyField(t *testing.T) {
	type Example struct {
		Name string `table:"name,omitempty"`
		Age  int    `table:"age,omitempty"`
	}

	given := Example{
		Name: "John",
	}

	want := []byte("name\n\"John\"\n")

	got, err := Marshal(given, WithFieldSeparator(","))
	if err != nil {
		t.Errorf("Marshal() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Marshal() = %v, want %v", string(got), string(want))
	}
	// opts := []cmp.Option{
	// 	cmp.AllowUnexported(structFields{}),
	// 	cmp.AllowUnexported(field{}),
	// 	cmpopts.IgnoreFields(field{}, "typ"),
	// 	cmpopts.IgnoreFields(field{}, "encoder"),
	// }

	// if diff := cmp.Diff(want, got, opts...); diff != "" {
	// 	t.Errorf("typeFields() mismatch (-want +got):\n%s", diff)
	// }
}

func TestNestedStruct(t *testing.T) {
	type Link struct {
		HRef string `json:"href,omitempty" yaml:"href,omitempty" table:"href,omitempty"`
		Name string `json:"name,omitempty" yaml:"name,omitempty" table:"name,omitempty"`
	}

	type Object struct {
		Self   *Link   `json:"self,omitempty" yaml:"self,omitempty" table:"self,omitempty"`
		HTML   *Link   `json:"html,omitempty" yaml:"html,omitempty" table:"html,omitempty"`
		Avatar *Link   `json:"avatar,omitempty" yaml:"avatar,omitempty" table:"avatar,omitempty"`
		Clone  []*Link `json:"clone,omitempty" yaml:"clone,omitempty" table:"clone,omitempty"`
	}

	type Repository struct {
		Links *Object `json:"links,omitempty" yaml:"links,omitempty" table:"links,omitempty"`
	}

	r := Repository{
		Links: &Object{
			Self: &Link{
				HRef: "https://self.com",
			},
			HTML: &Link{
				HRef: "https://html.com",
			},
			Avatar: &Link{
				HRef: "https://avatart.com",
			},
		},
	}

	want := []byte("links.self.href,links.html.href,links.avatar.href\n\"https://self.com\",\"https://html.com\",\"https://avatart.com\"\n")

	got, err := Marshal(r, WithFieldSeparator(","))
	if err != nil {
		t.Errorf("Marshal() error = %v", err)
	}

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

func TestFlattenArrayOnStruct(t *testing.T) {
	t.Run("FlattenArrayOnStruct", func(t *testing.T) {
		type Nested struct {
			List []string `table:"list"`
		}

		given := Nested{
			List: []string{"a", "b", "c"},
		}
		want := []byte("list[0],list[1],list[2]\n\"a\",\"b\",\"c\"\n")

		got, err := Marshal(given, WithFieldSeparator(","), WithFlattenArray(true))
		if err != nil {
			t.Errorf("Marshal() error = %v", err)
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("Marshal() = %v, want %v", string(got), string(want))
		}
	})

	t.Run("NoFlattenArrayOnStruct", func(t *testing.T) {
		type Nested struct {
			List []string `table:"list"`
		}

		given := Nested{
			List: []string{"a", "b", "c"},
		}
		want := []byte("list\n'[\"a\",\"b\",\"c\"]'\n")

		got, err := Marshal(given, WithFieldSeparator(","), WithFlattenArray(false))
		if err != nil {
			t.Errorf("Marshal() error = %v", err)
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("Marshal() = %v, want %v", string(got), string(want))
		}
	})
}

func TestFlattenArrayOnNestedStruct(t *testing.T) {
	type Level2 struct {
		List3 []string `table:"list3"`
	}
	type Level1 struct {
		List2 []Level2 `table:"list2"`
	}
	type Level0 struct {
		List1 []Level1 `table:"list1"`
	}

	given := Level0{
		List1: []Level1{
			{
				List2: []Level2{
					{
						List3: []string{"a", "b", "c"},
					},
					{
						List3: []string{"c", "d", "e"},
					},
				},
			},
			{
				List2: []Level2{
					{
						List3: []string{"a", "b", "c"},
					},
					{
						List3: []string{"c", "d", "e"},
					},
				},
			},
		},
	}
	header := []byte("list1.list2.list3[0],list1.list2.list3[1],list1.list2.list3[2]\n")
	row1 := []byte("\"a\",\"b\",\"c\"\n")
	row2 := []byte("\"c\",\"d\",\"e\"\n")
	row3 := []byte("\"a\",\"b\",\"c\"\n")
	row4 := []byte("\"c\",\"d\",\"e\"\n")

	want := append(header, row1...)
	want = append(want, row2...)
	want = append(want, row3...)
	want = append(want, row4...)

	got, err := Marshal(given, WithFieldSeparator(","), WithFlattenArray(true))
	if err != nil {
		t.Errorf("Marshal() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Marshal() = %v, want %v", string(got), string(want))
	}
}
