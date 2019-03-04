package sugar

import "github.com/graphql-go/graphql"

// the default arg loader defined here should work for most users, saving them from having to
// instantiate one themselves.  There are also convenience functions for working with the default
// loader.  Think of this as comparable to the DefaultServeMuxin the http package.

var defaultLoader *ArgLoader

func init() {
	var err error
	defaultLoader, err = New()
	if err != nil {
		panic(err)
	}
}

// ArgsConfig takes a struct instance with appropriate struct tags on its fields and returns a map
// of argument names to graphql argument configs, for assigning to the Args field in a
// graphql.Field.  If there is an error generating the argument configs, this function will panic.
// It uses the default arg loader.
func ArgsConfig(i interface{}) graphql.FieldConfigArgument {
	return defaultLoader.ArgsConfig(i)
}

// SafeArgsConfig -- this takes a struct instance w/ tags and returns a map of argument names.
// This will not cause a panic upon error.  It uses the default arg loader.
func SafeArgsConfig(i interface{}) (graphql.FieldConfigArgument, error) {
	return defaultLoader.SafeArgsConfig(i)
}

// RegisterArgParser takes a func (string) (<anytype>, error) and registers it on the ArgLoader as
// the parser for <anytype>.  It uses the default arg loader.
func RegisterArgParser(f interface{}, gqlType graphql.Output) error {
	return defaultLoader.RegisterArgParser(f, gqlType)
}

// LoadArgs loads arguments from the ResolveParam's map into the provided struct.  It uses the
// default arg loader.
func LoadArgs(p graphql.ResolveParams, c interface{}) error {
	return defaultLoader.LoadArgs(p, c)
}
