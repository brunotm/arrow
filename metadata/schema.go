package metadata

type Schema struct {
	fields      []Field
	nameToIndex map[string]int
	metadata    KeyValueMetadata
}

type Field struct {
	name     string           // Field name
	typ      DataType         // The field's data type
	nullable bool             // Fields can be nullable
	metadata KeyValueMetadata // The field's metadata, if any
}

type KeyValueMetadata struct {
	keys   []string
	values []string
}
