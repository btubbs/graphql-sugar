package sugar

import "github.com/graphql-go/graphql"

// the default type builder and convenience functions defined here should work for most users,
// saving them from having to instantiate one themselves.  Use this if you only have need for one
// set of types.  If you need multiple sets in the same Go process, then instantiate separate
// TypeBuilders.

var defaultTypeBuilder *TypeBuilder = NewTypeBuilder()

// RegisterKnownType takes any value, and the GraphQL type that should represent it, and will use that when building types.
func RegisterKnownType(val interface{}, gqlType graphql.Output) {
	defaultTypeBuilder.RegisterKnownType(val, gqlType)
}

// OutputType builds a GraphQL type for you from the given name, desc, and val.  val may be a
// struct instance, slice, array, pointer, or an alias to a builtin type like int or string.
func OutputType(name, desc string, val interface{}) graphql.Output {
	return defaultTypeBuilder.OutputType(name, desc, val)
}

// Union is just like OutputType, but takes in multiple vals and builds a GraphQL union type out of
// them.
func Union(name, desc string, vals ...interface{}) *graphql.Union {
	return defaultTypeBuilder.Union(name, desc, vals...)
}
