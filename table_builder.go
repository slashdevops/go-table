package table

import "io"

type tableBuilderChoice struct {
	t        *Table
	output   io.Writer
	sep      string
	minWidth int
	tabWidth int
	padding  int
	padChar  byte
	flags    uint
}

// NewBuilder creates a new table builder choice to create a table
func NewBuilder(output io.Writer) *tableBuilderChoice {
	return &tableBuilderChoice{
		t:      New(output),
		output: output,
	}
}

// WithHeader sets the header of the table
func (t *tableBuilderChoice) WithHeader(header []string) *tableBuilderChoice {
	t.t.SetHeader(header)
	return t
}

// WithRow adds a row to the table
func (t *tableBuilderChoice) WithRow(row ...string) *tableBuilderChoice {
	t.t.AddRow(row...)
	return t
}

// WithRowf adds a row to the table using a format string
func (t *tableBuilderChoice) WithRowf(format string, a ...interface{}) *tableBuilderChoice {
	t.t.AddRowf(format, a...)
	return t
}

// WithRows adds multiple rows to the table, matrix style
func (t *tableBuilderChoice) WithRows(rows [][]string) *tableBuilderChoice {
	t.t.AddRows(rows)
	return t
}

// WithRowsf adds multiple rows to the table, matrix style, using a format string
func (t *tableBuilderChoice) WithRowsf(format string, a ...interface{}) *tableBuilderChoice {
	t.t.AddRowsf(format, a...)
	return t
}

// WithSep sets the separator of the table
func (t *tableBuilderChoice) WithSep(sep string) *tableBuilderChoice {
	t.sep = sep
	return t
}

// WithMinWidth sets the minimum width of the table
func (t *tableBuilderChoice) WithMinWidth(minWidth int) *tableBuilderChoice {
	t.minWidth = minWidth
	return t
}

// WithTabWidth sets the tab width of the table
func (t *tableBuilderChoice) WithTabWidth(tabWidth int) *tableBuilderChoice {
	t.tabWidth = tabWidth
	return t
}

// WithPadding sets the padding of the table
func (t *tableBuilderChoice) WithPadding(padding int) *tableBuilderChoice {
	t.padding = padding
	return t
}

// WithPadChar sets the pad char of the table
func (t *tableBuilderChoice) WithPadChar(padChar byte) *tableBuilderChoice {
	t.padChar = padChar
	return t
}

// WithFlags sets the flags of the table
func (t *tableBuilderChoice) WithFlags(flags uint) *tableBuilderChoice {
	t.flags = flags
	return t
}

// Build builds the table
func (t *tableBuilderChoice) Build() *Table {
	opts := []TableOption{}
	if t.sep != "" {
		opts = append(opts, WithSep(t.sep))
	}
	if t.minWidth != 0 {
		opts = append(opts, WithMinWidth(t.minWidth))
	}
	if t.tabWidth != 0 {
		opts = append(opts, WithTabWidth(t.tabWidth))
	}
	if t.padding != 0 {
		opts = append(opts, WithPadding(t.padding))
	}
	if t.padChar != 0 {
		opts = append(opts, WithPadChar(t.padChar))
	}
	if t.flags != 0 {
		opts = append(opts, WithFlags(t.flags))
	}

	table := New(
		t.output,
		opts...,
	)

	if t.t.header != nil {
		table.SetHeader(t.t.header)
	}

	for _, row := range t.t.rows {
		table.AddRow(row...)
	}

	return table
}
