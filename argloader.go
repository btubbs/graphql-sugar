package sugar

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/graphql-go/graphql"
	multierror "github.com/hashicorp/go-multierror"
)

type tagKey string

const (
	defaultTag                = "arg"
	defaultSeparator          = ","
	defaultHidden             = "-"
	defaultAssignor           = ":"
	descTag                   = "desc"
	tagKeyRequired     tagKey = "required"
	tagKeyCoalesceZero tagKey = "coalesceZero"
)

// DefaultLoaders are for extra types beyond the 4 scalar types built into GraphQL.
var DefaultLoaders = []struct {
	LoaderFunc interface{}
	GqlType    graphql.Output
}{
	// Go std library types.  Keep these.
	{LoaderFunc: LoadRawJSON, GqlType: JSON},
	{LoaderFunc: LoadUInt, GqlType: graphql.Int},
}

// BaseLoaders are for the 4 scalar types built into GraphQL.
var BaseLoaders = []struct {
	LoaderFunc interface{}
	GqlType    graphql.Output
}{
	{LoaderFunc: LoadBool, GqlType: graphql.Boolean},
	{LoaderFunc: LoadBoolPointer, GqlType: graphql.Boolean},
	{LoaderFunc: LoadString, GqlType: graphql.String},
	{LoaderFunc: LoadInt, GqlType: graphql.Int},
	{LoaderFunc: LoadFloat, GqlType: graphql.Float},
	{LoaderFunc: LoadTime, GqlType: Timestamp},
}

// New returns a ArgLoader with all default loader funcs enabled.
func New() (*ArgLoader, error) {
	ec, err := Base()
	if err != nil {
		return nil, fmt.Errorf("could not load default arg funcs: %v", err)
	}
	for _, l := range DefaultLoaders {
		err = ec.RegisterArgParser(l.LoaderFunc, l.GqlType)
		if err != nil {
			return nil, err
		}
	}
	return ec, nil
}

// Base returns a ArgLoader with the 4 base loader funcs enabled.
func Base() (*ArgLoader, error) {
	ec := Empty()
	for _, l := range BaseLoaders {
		err := ec.RegisterArgParser(l.LoaderFunc, l.GqlType)
		if err != nil {
			return nil, err
		}
	}
	return ec, nil
}

// Empty returns a ArgLoader without any loader funcs enabled.
func Empty() *ArgLoader {
	ec := &ArgLoader{}
	ec.loaderFuncs = map[reflect.Type]func(interface{}, map[tagKey]string) (reflect.Value, error){}
	ec.gqlTypes = map[reflect.Type]graphql.Output{}
	return ec
}

// ArgLoader is a helper for reading arguments from a graphql.ResolveParams, converting them to Go
// types, and setting their values to fields on a user-provided struct.
type ArgLoader struct {
	// a map from reflect types to functions that can take an interface and return a
	// reflect value of that type.
	loaderFuncs map[reflect.Type]func(interface{}, map[tagKey]string) (reflect.Value, error)

	// a map from reflect types to the graphql types that should be used for their arguments.
	gqlTypes map[reflect.Type]graphql.Output
}

// ArgsConfig takes a struct instance with appropriate struct tags on its fields and returns a map
// of argument names to graphql argument configs, for assigning to the Args field in a
// graphql.Field.  If there is an error generating the argument configs, this function will panic.
func (e *ArgLoader) ArgsConfig(i interface{}) graphql.FieldConfigArgument {
	conf, err := e.SafeArgsConfig(i)
	if err != nil {
		panic(fmt.Sprintf("could not configure arguments: %v", err))
	}
	return conf
}

func readTag(field reflect.StructField) (string, map[tagKey]string, bool) {
	v, ok := field.Tag.Lookup(defaultTag)
	if !ok {
		// this field doesn't have our tag.  Skip.
		return "", nil, false
	}
	values := strings.Split(v, defaultSeparator)
	if len(values) < 1 || values[0] == defaultHidden {
		return "", nil, false
	}
	name := values[0]
	if name == "" {
		name = field.Name
	}
	config := map[tagKey]string{}
	for _, value := range values {
		keyValuePair := strings.SplitN(value, defaultAssignor, 2)
		if len(keyValuePair) < 1 {
			return "", nil, false
		} else if len(keyValuePair) < 2 {
			config[tagKey(keyValuePair[0])] = ""
		} else {
			config[tagKey(keyValuePair[0])] = keyValuePair[1]
		}
	}
	return name, config, true
}

// SafeArgsConfig -- this takes a struct instance w/ tags and returns a map of argument names.
// This will not cause a panic upon error
func (e *ArgLoader) SafeArgsConfig(i interface{}) (graphql.FieldConfigArgument, error) {
	// we should have a struct
	var structType reflect.Type

	// accept either a struct or a pointer to a struct
	iType := reflect.TypeOf(i)
	if iType.Kind() == reflect.Ptr {
		structType = iType.Elem()
	} else {
		structType = reflect.TypeOf(i)
	}
	if structType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%v is not a struct", i)
	}

	out := graphql.FieldConfigArgument{}
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		argName, _, ok := readTag(field)
		if !ok {
			// this field doesn't have our tag.  Skip.
			continue
		}

		if argType, ok := e.gqlTypes[field.Type]; ok {
			out[argName] = &graphql.ArgumentConfig{
				Type:        argType,
				Description: field.Tag.Get(descTag),
			}
		} else {
			return nil, fmt.Errorf("no argument loader registered for %v type", field.Type)
		}

	}
	return out, nil
}

// RegisterArgParser takes a func (string) (<anytype>, error) and registers it on the ArgLoader as
// the parser for <anytype>
func (e *ArgLoader) RegisterArgParser(f interface{}, gqlType graphql.Output) error {
	// alright, let's inspect this f and make sure it's a func (string) (sometype, err)
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Func {
		return fmt.Errorf("%v is not a func", f)
	}

	fname := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	// f should accept one argument
	if t.NumIn() != 1 {
		return fmt.Errorf(
			"loader func should accept 1 interface{} argument. %v accepts %d arguments",
			fname, t.NumIn())
	}
	// it should return two things
	if t.NumOut() != 2 {
		return fmt.Errorf(
			"loader func should return 2 arguments. %v returns %d arguments",
			fname, t.NumOut())
	}
	// the first can be any type. the second should be error
	errorInterface := reflect.TypeOf((*error)(nil)).Elem()
	if !t.Out(1).Implements(errorInterface) {
		return fmt.Errorf(
			"loader func's last return value should be error. %s's last return value is %v",
			fname, t.Out(1))
	}
	_, alreadyRegistered := e.loaderFuncs[t.Out(0)]
	if alreadyRegistered {
		return fmt.Errorf("a loader func has already been registered for the %v type.  cannot also register %s",
			t.Out(0), fname,
		)
	}

	callable := reflect.ValueOf(f)
	wrapped := func(i interface{}, config map[tagKey]string) (v reflect.Value, err error) {
		defer func() {
			if p := recover(); p != nil {
				// we panicked running the inner loader func.
				err = fmt.Errorf("%s panicked: %s", fname, p)
			}
		}()
		returnvals := callable.Call([]reflect.Value{reflect.ValueOf(i)})
		// check for non nil error
		if !returnvals[1].IsNil() {
			return reflect.Value{}, fmt.Errorf("%v", returnvals[1])
		}
		return returnvals[0], nil
	}
	e.loaderFuncs[t.Out(0)] = wrapped
	e.gqlTypes[t.Out(0)] = gqlType
	return nil
}

// LoadArgs loads arguments from the provided map into the provided struct.
func (e *ArgLoader) LoadArgs(p graphql.ResolveParams, c interface{}) error {
	// assert that c is a struct.
	cType := reflect.TypeOf(c)
	cVal := reflect.ValueOf(c)
	if cType.Kind() == reflect.Ptr {
		cType = cType.Elem()
		cVal = cVal.Elem()
	}

	if cType.Kind() != reflect.Struct {
		return fmt.Errorf("%v is not a struct", c)
	}

	valErrs := multierror.Append(nil)
	for i := 0; i < cType.NumField(); i++ {
		field := cType.Field(i)

		argName, config, ok := readTag(field)
		if !ok {
			// this field doesn't have our tag.  Skip.
			continue
		}

		interfaceVal, ok := p.Args[argName]
		if !ok {
			// could not find the key we're looking for in map.  is it required?
			if _, ok := config[tagKeyRequired]; ok {
				multierror.Append(valErrs, fmt.Errorf("%s is required", argName))
			}
			continue
		}
		loaderFunc, ok := e.loaderFuncs[field.Type]
		if !ok {
			return fmt.Errorf("no loader function found for type %v", field.Type)
		}

		toSet, err := loaderFunc(interfaceVal, config)
		if err != nil {
			if _, ok := config[tagKeyCoalesceZero]; !ok {
				valErrs = multierror.Append(valErrs, fmt.Errorf("%s is not valid", argName))
				continue
			}
			toSet = reflect.Zero(field.Type)
		}
		cVal.Field(i).Set(toSet)
	}
	return valErrs.ErrorOrNil()
}
