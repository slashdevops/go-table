package table

import (
	"fmt"
	"io"
	"runtime"
	"strings"
	"text/tabwriter"
)

const (
	// this is the tag name used to identify fields in a struct
	structTagName = "table"

	// this is the default header name used when a simple go value is passed to the table
	UnknownHeaderName = "Unknown"
)

const (
	// Ignore html tags and treat entities (starting with '&'
	// and ending in ';') as single characters (width = 1).
	FilterHTML uint = 1 << iota

	// Strip Escape characters bracketing escaped text segments
	// instead of passing them through unchanged with the text.
	StripEscape

	// Force right-alignment of cell content.
	// Default is left-alignment.
	AlignRight

	// Handle empty columns as if they were not present in
	// the input in the first place.
	DiscardEmptyColumns

	// Always use tabs for indentation columns (i.e., padding of
	// leading empty cells on the left) independent of padchar.
	TabIndent

	// Print a vertical bar ('|') between columns (after formatting).
	// Discarded columns appear as zero-width columns ("||").
	Debug
)

var carriageReturn string = "\n"

// Table is a simple abstraction for tex/tabwriter golang package
type Table struct {
	tw     *tabwriter.Writer
	rows   [][]string
	header []string
	sep    string
}

// New creates a new table
func New(output io.Writer, opts ...TableOption) *Table {
	if runtime.GOOS == "windows" {
		carriageReturn = "\r\n"
	}

	o := &tableOptions{
		sep:      "\t",
		minWidth: 0,
		tabWidth: 0,
		padding:  2,
		padChar:  ' ',
		flags:    0,
	}
	for _, opt := range opts {
		opt(o)
	}

	return &Table{
		tw:     tabwriter.NewWriter(output, o.minWidth, o.tabWidth, o.padding, o.padChar, o.flags),
		rows:   make([][]string, 0),
		header: make([]string, 0),
		sep:    o.sep,
	}
}

// SetHeader sets the header of the table
func (t *Table) SetHeader(header []string) {
	t.header = header
}

// AddRow adds a row to the table
func (t *Table) AddRow(row ...string) {
	t.rows = append(t.rows, row)
}

// AddRowf adds a row to the table using a format string
func (t *Table) AddRowf(format string, a ...any) {
	t.rows = append(t.rows, []string{fmt.Sprintf(format, a...)})
}

// AddRows adds multiple rows to the table, matrix style
func (t *Table) AddRows(rows [][]string) {
	t.rows = append(t.rows, rows...)
}

// AddRowsf adds multiple rows to the table, matrix style, using a format string
func (t *Table) AddRowsf(format string, a ...any) {
	for _, v := range a {
		t.rows = append(t.rows, []string{fmt.Sprintf(format, v)})
	}
}

// headerRow returns the header row
func (t *Table) headerRow() string {
	return row(t.sep, t.header...)
}

// row returns a row
func row(sep string, row ...string) string {
	return strings.Join(row, sep) + carriageReturn
}

// Render renders the table into the output
func (t *Table) Render() error {
	if len(t.header) > 0 {
		if _, err := t.tw.Write([]byte(t.headerRow())); err != nil {
			return err
		}
	}

	for _, r := range t.rows {
		if t.header == nil || len(t.header) == 0 {
			for j := range r {
				t.header = append(t.header, fmt.Sprintf("%s%d", UnknownHeaderName, j+1))
			}
			if _, err := t.tw.Write([]byte(t.headerRow())); err != nil {
				return err
			}
		}

		if _, err := t.tw.Write([]byte(row(t.sep, r...))); err != nil {
			return err
		}
	}

	return t.tw.Flush()
}
