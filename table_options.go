package table

// minWidth	minimal cell width including any padding
// tabWidth	width of tab characters (equivalent number of spaces)
// padding		padding added to a cell before computing its width
// padChar		ASCII char used for padding
//
//	if padChar == '\t', the Writer will assume that the
//	width of a '\t' in the formatted output is tabWidth,
//	and cells are left-aligned independent of align_left
//	(for correct-looking results, tabWidth must correspond
//	to the tab width in the viewer displaying the result)
//
// flags		formatting control
type tableOptions struct {
	sep      string
	minWidth int
	tabWidth int
	padding  int
	padChar  byte
	flags    uint
}

// TableOption is a function that configures a table
type TableOption func(*tableOptions)

// WithSep sets the separator of the table
func WithSep(sep string) TableOption {
	return func(o *tableOptions) {
		o.sep = sep
	}
}

// WithMinWidth sets the minimum width of the table
func WithMinWidth(minWidth int) TableOption {
	return func(o *tableOptions) {
		o.minWidth = minWidth
	}
}

// WithTabWidth sets the tab width of the table
func WithTabWidth(tabWidth int) TableOption {
	return func(o *tableOptions) {
		o.tabWidth = tabWidth
	}
}

// WithPadding sets the padding of the table
func WithPadding(padding int) TableOption {
	return func(o *tableOptions) {
		o.padding = padding
	}
}

// WithPadChar sets the pad char of the table
func WithPadChar(padChar byte) TableOption {
	return func(o *tableOptions) {
		o.padChar = padChar
	}
}

// WithFlags sets the flags of the table
func WithFlags(flags uint) TableOption {
	return func(o *tableOptions) {
		o.flags = flags
	}
}
