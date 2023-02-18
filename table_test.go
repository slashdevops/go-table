package table

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"testing"
	"text/tabwriter"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		output     io.Writer
		opts       []TableOption
		want       *Table
		wantOutput string
	}{
		{
			name:   "default",
			output: &bytes.Buffer{},
			opts:   []TableOption{},
			want: &Table{
				tw:     tabwriter.NewWriter(&bytes.Buffer{}, 0, 0, 2, ' ', 0),
				rows:   make([][]any, 0),
				header: make([]string, 0),
				sep:    "\t",
			},
			wantOutput: "",
		},
		{
			name:   "with options",
			output: &bytes.Buffer{},
			opts: []TableOption{
				WithMinWidth(10),
				WithTabWidth(10),
				WithPadding(10),
				WithPadChar('|'),
				WithFlags(AlignRight | Debug),
				WithSep("\t"),
			},
			want: &Table{
				tw:     tabwriter.NewWriter(&bytes.Buffer{}, 10, 10, 10, '|', AlignRight|Debug),
				rows:   make([][]any, 0),
				header: make([]string, 0),
				sep:    "\t",
			},
			wantOutput: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			if got := New(output, tt.opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
			if gotOutput := output.String(); gotOutput != tt.wantOutput {
				t.Errorf("New() = %v, want %v", gotOutput, tt.wantOutput)
			}
		})
	}
}

// ***********************************************************************************************
// ************************* Examples ************************************************************

// ExampleNew shows how to create a new table
func ExampleNew() {
	buf := new(bytes.Buffer)
	t := New(buf)
	t.SetHeader([]string{"Name", "Age", "City"})
	t.AddRow([]any{"John", "30", "New York"}...)
	t.AddRow([]any{"Jane", "20", "London"}...)
	t.AddRow([]any{"Jack", "40", "Paris"}...)
	t.Render()

	fmt.Println(buf.String())
	// Output:
	// Name  Age  City
	// John  30   New York
	// Jane  20   London
	// Jack  40   Paris
}

// ExampleNew shows how to create a new table with options
func ExampleNew_withOptions() {
	buf := new(bytes.Buffer)
	t := New(
		buf,
		WithMinWidth(10),
		WithTabWidth(10),
		WithPadding(10),
		WithPadChar('.'),
		WithFlags(AlignRight|Debug),
		WithSep("\t"),
	)
	t.SetHeader([]string{"Name", "Age", "City"})
	t.AddRow([]any{"John", "30", "New York"}...)
	t.AddRow([]any{"Jane", "20", "London"}...)
	t.AddRow([]any{"Jack", "40", "Paris"}...)
	t.Render()

	fmt.Println(buf.String())
	// Output:
	// ..........Name|..........Age|City
	// ..........John|...........30|New York
	// ..........Jane|...........20|London
	// ..........Jack|...........40|Paris
}

// ExampleTable_SetHeader_noCall show the default behavior when no header is set
func ExampleTable_SetHeader_noCall() {
	buf := new(bytes.Buffer)
	t := New(buf)
	// t.SetHeader([]string{"Name", "Age", "City"}) // with no headers, show Unknown[1,2,3...]
	t.AddRow([]any{"John", "30", "New York"}...)
	t.AddRow([]any{"Jane", "20", "London"}...)
	t.AddRow([]any{"Jack", "40", "Paris"}...)
	t.Render()

	fmt.Println(buf.String())
	// Output:
	// Unknown1  Unknown2  Unknown3
	// John      30        New York
	// Jane      20        London
	// Jack      40        Paris
}

// ExampleNew_withOptions_csv shows how to output a csv
func ExampleNew_withOptions_csv() {
	buf := new(bytes.Buffer)
	t := New(buf, WithSep(","))
	t.SetHeader([]string{"Name", "Age", "City"})
	t.AddRow([]any{"John", "30", "New York"}...)
	t.AddRow([]any{"Jane", "20", "London"}...)
	t.AddRow([]any{"Jack", "40", "Paris"}...)
	t.Render()

	fmt.Println(buf.String())
	// Output:
	// Name,Age,City
	// John,30,New York
	// Jane,20,London
	// Jack,40,Paris
}

// ExampleTable_builder shows how tou use the builder pattern and regular methods
func ExampleTable_builder() {
	buf := new(bytes.Buffer)
	t := NewBuilder(buf).
		WithMinWidth(10).
		WithTabWidth(10).
		WithPadding(10).
		WithPadChar('.').
		WithFlags(AlignRight | Debug).
		WithSep("\t").
		WithHeader([]string{"Name", "Age", "City"}).
		WithRow([]any{"John", "30", "New York"}...).
		WithRow([]any{"Jane", "20", "London"}...).
		Build()

	t.AddRow([]any{"Jack", "40", "Paris"}...)

	t.Render()

	fmt.Println(buf.String())
	// Output:
	// ..........Name|..........Age|City
	// ..........John|...........30|New York
	// ..........Jane|...........20|London
	// ..........Jack|...........40|Paris
}

// ExampleTable_builder2 shows how tou use the builder pattern and regular methods
// in different order and code style
func ExampleTable_builder2() {
	out := new(bytes.Buffer)
	t := NewBuilder(out).
		WithFlags(Debug)

		// add header and rows before building
	t.WithHeader([]string{"Name", "Age", "City"})
	t.WithRow([]any{"John", "30", "New York"}...)
	t.WithRow([]any{"Jane", "20", "London"}...)
	t.WithRow([]any{"Jack", "40", "Paris"}...)
	t.WithRow([]any{"Christian", "47", "Barcelona"}...)

	// build the table
	table := t.Build()

	// add a new row after building
	table.AddRow([]any{"Ely", "50", "Barcelona"}...)

	// render the table
	table.Render()

	fmt.Println(out.String())
	// Output:
	// Name       |Age  |City
	// John       |30   |New York
	// Jane       |20   |London
	// Jack       |40   |Paris
	// Christian  |47   |Barcelona
	// Ely        |50   |Barcelona
}
