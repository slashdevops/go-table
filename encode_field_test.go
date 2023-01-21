package table

import (
	"reflect"
	"testing"
)

func Test_typeFields(t *testing.T) {
	type sc struct {
		ScField1 string `table:"scField1"`
	}

	type sb struct {
		SbField1 sc `table:"sbField1"`
	}

	type sa struct {
		SaField1 string  `table:"saField1"`
		SaField2 int     `table:"saField2"`
		SaField3 sb      `table:"saField3"`
		SaField4 float32 `table:"saField4"`
		SaField5 sc      `table:"saField5"`
	}

	s := sa{
		SaField1: "value of field1",
		SaField2: 1,
		SaField3: sb{
			SbField1: sc{
				ScField1: "value of field4",
			},
		},
		SaField4: 1.1,
		SaField5: sc{
			ScField1: "value of field5",
		},
	}

	tests := []struct {
		name string
		t    reflect.Type
		want structFields
	}{
		{
			name: "test",
			t:    reflect.TypeOf(s),
			want: structFields{
				list: []field{
					{
						name:      "saField1",
						nameBytes: []byte("saField1"),
						index:     []int{0},
						tag:       true,
						typ:       reflect.TypeOf("string"),
						omitEmpty: false,
						quoted:    false,
						encoder:   typeEncoder(reflect.TypeOf("string")),
					},
					{
						name:      "saField2",
						nameBytes: []byte("saField2"),
						index:     []int{1},
						tag:       true,
						typ:       reflect.TypeOf(int(1)),
						omitEmpty: false,
						quoted:    false,
						encoder:   typeEncoder(reflect.TypeOf(int(1))),
					},
					{
						name:      "saField3.sbField1.scField1",
						nameBytes: []byte("saField3.sbField1.scField1"),
						index:     []int{2, 0, 0},
						tag:       true,
						typ:       reflect.TypeOf("string"),
						omitEmpty: false,
						quoted:    false,
						encoder:   typeEncoder(reflect.TypeOf("string")),
					},
					{
						name:      "saField4",
						nameBytes: []byte("saField4"),
						index:     []int{3},
						tag:       true,
						typ:       reflect.TypeOf(float32(1.1)),
						omitEmpty: false,
						quoted:    false,
						encoder:   typeEncoder(reflect.TypeOf(float32(1.1))),
					},
					{
						name:      "saField5.scField1",
						nameBytes: []byte("saField5.scField1"),
						index:     []int{4, 0},
						tag:       true,
						typ:       reflect.TypeOf("string"),
						omitEmpty: false,
						quoted:    false,
						encoder:   typeEncoder(reflect.TypeOf("string")),
					},
				},
				nameIndex: map[string]int{
					"saField1": 0, "saField2": 1, "saField3.sbField1.scField1": 2, "saField4": 3, "saField5.scField1": 4,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := typeFields(tt.t, "")

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("typeFields() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
