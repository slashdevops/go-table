package table

type marshalOptions struct {
	sep          string
	flattenArray bool
	flattenMap   bool
}

// EncodeOptions is a function that configures encode
type MarshalOption func(*marshalOptions)

// WithFieldSeparator sets the separator of the table
func WithFieldSeparator(sep string) MarshalOption {
	return func(o *marshalOptions) {
		o.sep = sep
	}
}

// WithFlattenArray sets the flatten array of the table
func WithFlattenArray(flattenArray bool) MarshalOption {
	return func(o *marshalOptions) {
		o.flattenArray = flattenArray
	}
}

// WithFlattenMap sets the flatten map of the table
func WithFlattenMap(flattenMap bool) MarshalOption {
	return func(o *marshalOptions) {
		o.flattenMap = flattenMap
	}
}
