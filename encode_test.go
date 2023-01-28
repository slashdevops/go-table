package table

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
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
		Field3 float64 `table:"float"`
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
		sep  []string
		want []byte
	}{
		{v: nil, sep: []string{","}, want: []byte("Unknown1\nnull\n")},
		{v: true, sep: []string{","}, want: []byte("Unknown1\ntrue\n")},
		{v: false, sep: []string{","}, want: []byte("Unknown1\nfalse\n")},
		{v: 12, sep: []string{","}, want: []byte("Unknown1\n12\n")},
		{v: uint64(111), sep: []string{","}, want: []byte("Unknown1\n111\n")},
		{v: float32(32.32), sep: []string{","}, want: []byte("Unknown1\n32.32\n")},
		{v: float64(64.64), sep: []string{","}, want: []byte("Unknown1\n64.64\n")},
		{v: "this is a string", sep: []string{","}, want: []byte("Unknown1\n\"this is a string\"\n")},
		{v: [5]int{1, 2, 3, 4, 5}, sep: []string{","}, want: []byte("Unknown1\n'[1,2,3,4,5]'\n")},
		{v: []int{1, 2, 3, 4, 5}, sep: []string{","}, want: []byte("Unknown1\n'[1,2,3,4,5]'\n")},
		{v: []float64{1.5, 2.5, 3.5, 4.5, 5.5}, sep: []string{","}, want: []byte("Unknown1\n'[1.5,2.5,3.5,4.5,5.5]'\n")},
		{v: []string{"this is a string", "b", "c"}, sep: []string{","}, want: []byte("Unknown1\n'[\"this is a string\",\"b\",\"c\"]'\n")},
		{v: map[string]int{"foo": 1, "bar": 2}, sep: []string{","}, want: []byte("Unknown1\n'{\"bar\":2,\"foo\":1}'\n")},
		{v: map[int]string{1: "foo", 2: "bar"}, sep: []string{","}, want: []byte("Unknown1\n'{\"1\":\"foo\",\"2\":\"bar\"}'\n")},
		{v: &[5]int{1, 2, 3, 4, 5}, sep: []string{","}, want: []byte("Unknown1\n'[1,2,3,4,5]'\n")},
		{
			v:    &st{A: 1, B: "this is a string", C: &sp, D: []string{"one", "two"}, E: map[string]string{"foo": "one", "bar": "two"}},
			sep:  []string{","},
			want: []byte("a,b,c,d,e\n1,\"this is a string\",\"pointer to string\",'[\"one\",\"two\"]','{\"bar\":\"two\",\"foo\":\"one\"}'\n"),
		},
		{
			v:    cmp,
			sep:  []string{","},
			want: []byte("field1,float,Field5,field6.a,field6.b,field6.c,field6.d,field6.e\n\"field1\",3.3,\"pointer to string\",1,\"b\",\"pointer to string\",'[\"d1\",\"d2\"]','{\"e1\":\"e1\",\"e2\":\"e2\"}'\n"),
		},
		{
			v:    cmp,
			sep:  []string{","},
			want: []byte("field1,float,Field5,field6.a,field6.b,field6.c,field6.d,field6.e\n\"field1\",3.3,\"pointer to string\",1,\"b\",\"pointer to string\",'[\"d1\",\"d2\"]','{\"e1\":\"e1\",\"e2\":\"e2\"}'\n"),
		},
	}

	for _, tt := range tests {
		var name string

		if tt.v == nil {
			name = "nil"
		} else {
			name = reflect.TypeOf(tt.v).String()
		}

		t.Run(name, func(t *testing.T) {
			got, err := Marshal(tt.v, tt.sep...)
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Marshal() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

// func TestMarshalRecursiveStruct(t *testing.T) {
// 	type st struct {
// 		Field1 string `table:"field1"`
// 		Field3 *st    `table:"field3"`
// 	}

// 	st1 := st{
// 		Field3: &st{
// 			Field1: "field1",
// 		},
// 	}

// 	want := []byte("a,b,c,d,e,f.a,f.b,f.c,f.d,f.e\n1,\"this is a string\",\"pointer to string\",'[\"one\",\"two\"]','{\"bar\":\"two\",\"foo\":\"one\"}',1,\"this is a string\",\"pointer to string\",'[\"one\",\"two\"]','{\"bar\":\"two\",\"foo\":\"one\"}'\n")

// 	got, err := Marshal(st1, ",")
// 	if err != nil {
// 		t.Errorf("Marshal() error = %v", err)
// 	}

// 	if !reflect.DeepEqual(got, want) {
// 		t.Errorf("Marshal() = %+v, want %+v", string(got), string(want))
// 	}
// }

func TestNestedStruct(t *testing.T) {
	type Link struct {
		HRef string `json:"href,omitempty" yaml:"href,omitempty" table:"href,omitempty"`
		Name string `json:"name,omitempty" yaml:"name,omitempty" table:"name,omitempty"`
	}

	type Object struct {
		Self         *Link   `json:"self,omitempty" yaml:"self,omitempty" table:"self,omitempty"`
		HTML         *Link   `json:"html,omitempty" yaml:"html,omitempty" table:"html,omitempty"`
		Avatar       *Link   `json:"avatar,omitempty" yaml:"avatar,omitempty" table:"avatar,omitempty"`
		PullRequests *Link   `json:"pullrequests,omitempty" yaml:"pullrequests,omitempty" table:"pullrequests,omitempty"`
		Commits      *Link   `json:"commits,omitempty" yaml:"commits,omitempty" table:"commits,omitempty"`
		Forks        *Link   `json:"forks,omitempty" yaml:"forks,omitempty" table:"forks,omitempty"`
		Watchers     *Link   `json:"watchers,omitempty" yaml:"watchers,omitempty" table:"watchers,omitempty"`
		Downloads    *Link   `json:"downloads,omitempty" yaml:"downloads,omitempty" table:"downloads,omitempty"`
		Clone        []*Link `json:"clone,omitempty" yaml:"clone,omitempty" table:"clone,omitempty"`
		Hooks        *Link   `json:"hooks,omitempty" yaml:"hooks,omitempty" table:"hooks,omitempty"`
	}

	type AccountLinks struct {
		Avatar *Link `json:"avatar,omitempty" yaml:"avatar,omitempty" table:"avatar,omitempty"`
	}

	type Account struct {
		Links       *AccountLinks `json:"links,omitempty" yaml:"links,omitempty" table:"links,omitempty"`
		CreatedOn   *time.Time    `json:"created_on,omitempty" yaml:"created_on,omitempty" table:"created_on,omitempty"`
		DisplayName string        `json:"display_name,omitempty" yaml:"display_name,omitempty" table:"display_name,omitempty"`
		Username    string        `json:"username,omitempty" yaml:"username,omitempty" table:"username,omitempty"`
		UUID        *uuid.UUID    `json:"uuid,omitempty" yaml:"uuid,omitempty" table:"uuid,omitempty"`
	}

	type Project struct {
		Links                   *Object    `json:"links,omitempty" yaml:"links,omitempty" table:"links,omitempty"`
		UUID                    *uuid.UUID `json:"uuid,omitempty" yaml:"uuid,omitempty" table:"uuid,omitempty"`
		Key                     string     `json:"key,omitempty" yaml:"key,omitempty" table:"key,omitempty"`
		Owner                   *Account   `json:"owner,omitempty" yaml:"owner,omitempty" table:"owner,omitempty"`
		Name                    string     `json:"name,omitempty" yaml:"name,omitempty" table:"name,omitempty"`
		Description             string     `json:"description,omitempty" yaml:"description,omitempty" table:"description,omitempty"`
		IsPrivate               bool       `json:"is_private,omitempty" yaml:"is_private,omitempty" table:"is_private,omitempty"`
		CreatedOn               *time.Time `json:"created_on,omitempty" yaml:"created_on,omitempty" table:"created_on,omitempty"`
		UpdatedOn               *time.Time `json:"updated_on,omitempty" yaml:"updated_on,omitempty" table:"updated_on,omitempty"`
		HasPubliclyVisibleRepos bool       `json:"has_publicly_visible_repos,omitempty" yaml:"has_publicly_visible_repos,omitempty" table:"has_publicly_visible_repos,omitempty"`
	}

	type Branch struct {
		MergeStrategies      []string `json:"merge_strategies,omitempty" yaml:"merge_strategies,omitempty" table:"merge_strategies,omitempty"`
		DefaultMergeStrategy string   `json:"default_merge_strategy,omitempty" yaml:"default_merge_strategy,omitempty" table:"default_merge_strategy,omitempty"`
	}

	type Repository struct {
		Links *Object `json:"links,omitempty" yaml:"links,omitempty" table:"links,omitempty"`
	}

	r := Repository{
		Links: &Object{
			Self: &Link{
				HRef: "https://api.bitbucket.org/2.0/repositories/username/repo",
			},
			HTML: &Link{
				HRef: "https://bitbucket.org/username/repo",
			},
			Avatar: &Link{
				HRef: "https://bitbucket.org/username/repo/avatar/32/",
			},
			Clone: []*Link{
				{
					HRef: "https://bitbucket.org/username/repo.git",
					Name: "https",
				},
			},
		},
	}

	want := []byte("a,b,c,d,e,f.a,f.b,f.c,f.d,f.e\n1,\"this is a string\",\"pointer to string\",'[\"one\",\"two\"]','{\"bar\":\"two\",\"foo\":\"one\"}',1,\"this is a string\",\"pointer to string\",'[\"one\",\"two\"]','{\"bar\":\"two\",\"foo\":\"one\"}'\n")

	got, err := Marshal(r, ",")
	if err != nil {
		t.Errorf("Marshal() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Marshal() = %+v, want %+v", string(got), string(want))
	}
}
