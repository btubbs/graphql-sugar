package sugar

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/guregu/null"
	"github.com/lib/pq"
)

// NewTypeBuilder creates a new TypeBuilder and registers known types on it.
func NewTypeBuilder() *TypeBuilder {
	tb := TypeBuilder{
		knownTypes: map[reflect.Type]graphql.Output{},
	}
	tb.RegisterKnownType(time.Now(), Timestamp)
	tb.RegisterKnownType(sql.NullString{}, graphql.String)
	tb.RegisterKnownType(null.Int{}, graphql.Int)
	tb.RegisterKnownType(null.String{}, graphql.String)
	tb.RegisterKnownType(pq.NullTime{}, Timestamp)
	tb.RegisterKnownType(null.Time{}, Timestamp)

	return &tb
}

// A TypeBuilder helps create graphql-go output types
type TypeBuilder struct {
	knownTypes map[reflect.Type]graphql.Output
}

// RegisterKnownType takes any value, and the GraphQL type that should represent it, and will use that when building types.
func (tb *TypeBuilder) RegisterKnownType(val interface{}, gqlType graphql.Output) {
	name := gqlType.Name()
	// panic if asked to create a graphql type with a lowercase first character.
	if string(name[0]) == strings.ToLower(string(name[0])) {
		panic(fmt.Sprintf("refusing to build GraphQL type with lowercase name %s. val: %+v", name, val))
	}
	tb.knownTypes[getType(val)] = gqlType
}

// OutputType builds a GraphQL type for you from the given name, desc, and val.  val may be a
// struct instance, slice, array, pointer, or an alias to a builtin type like int or string.
func (tb *TypeBuilder) OutputType(name, desc string, val interface{}) graphql.Output {
	// obj can be a reflect.type, or a concrete value
	objType := getType(val)

	// check known types first, so we don't recurse into time.Time structs, for example.
	if knownType, ok := tb.knownTypes[objType]; ok {
		return knownType
	}

	switch kind := objType.Kind(); kind {
	case reflect.Bool:
		tb.RegisterKnownType(objType, graphql.Boolean)
		return graphql.Boolean
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		tb.RegisterKnownType(objType, graphql.Int)
		return graphql.Int
	case reflect.Float32, reflect.Float64:
		tb.RegisterKnownType(objType, graphql.Float)
		return graphql.Float
	case reflect.String:
		tb.RegisterKnownType(objType, graphql.String)
		return graphql.String
	case reflect.Ptr:
		t := tb.OutputType(name, desc, objType.Elem())
		tb.RegisterKnownType(objType, t)
		return t
	case reflect.Slice, reflect.Array:
		t := graphql.NewList(tb.OutputType(name, desc, objType.Elem()))
		tb.RegisterKnownType(objType, t)
		return t
	case reflect.Struct:
		obj := graphql.NewObject(graphql.ObjectConfig{
			Name:        name,
			Description: desc,
			Fields:      tb.structFieldMap(objType, false),
		})
		// remember this type in our map to keep things consistent and fast.
		tb.RegisterKnownType(objType, obj)
		return obj
	default:
		fmt.Println("type", objType, "name", name)
		panic(fmt.Sprintf("cannot convert %v kind", objType.Kind()))
	}
}

// Union is just like OutputType, but takes in multiple vals and builds a GraphQL union type out of
// them.
func (tb *TypeBuilder) Union(name, desc string, vals ...interface{}) *graphql.Union {
	// a map to be used in the type resolver
	typeMap := map[reflect.Type]*graphql.Object{}

	// a list to be fed into the type definition
	typeList := []*graphql.Object{}

	for _, v := range vals {
		objType := getType(v)
		if gqlType, ok := tb.knownTypes[objType]; ok {
			gqlObj := gqlType.(*graphql.Object)
			typeMap[objType] = gqlObj
			typeList = append(typeList, gqlObj)
		} else {
			panic(fmt.Sprintf("must register a type for %v before using it in a union type", v))
		}
	}
	return graphql.NewUnion(graphql.UnionConfig{
		Name:        name,
		Description: desc,
		Types:       typeList,
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			typ := reflect.Indirect(reflect.ValueOf(p.Value)).Type()
			// if typ is a pointer type, get its elem.
			if outType, ok := typeMap[typ]; ok {
				return outType
			}
			panic(fmt.Sprintf("could not determine GraphQL type for %v", typ))
		},
	})
}

func (tb *TypeBuilder) structFieldMap(val interface{}, embedded bool) graphql.Fields {
	structType := getType(val)
	fieldMap := graphql.Fields{}
	// loop over fields on struct.
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		jsonName := field.Tag.Get("json")

		if field.Anonymous {
			// an embedded struct should be flattened
			embeddedFields := tb.structFieldMap(field.Type, true)
			for k, v := range embeddedFields {
				// don't overwrite fields already present from the parent
				if _, ok := fieldMap[k]; !ok {
					fieldMap[k] = v
				}
			}
		}
		switch jsonName {
		case "-": //skip
			continue
		case "":
			continue
		default:
			gqlField := &graphql.Field{
				Type:        tb.OutputType(jsonName, field.Tag.Get("desc"), field.Type),
				Description: field.Tag.Get("desc"),
			}

			// if there's a "deprecation" tag on the struct field, add DeprecationReason to gql field.
			if deprecation, ok := field.Tag.Lookup("deprecation"); ok {
				gqlField.DeprecationReason = deprecation
			}

			// if we're in an embedded struct, then we need to add a resolver
			if embedded {
				gqlField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
					val := reflect.ValueOf(p.Source)
					return reflect.Indirect(val).FieldByName(field.Name).Interface(), nil
				}
			}
			fieldMap[jsonName] = gqlField
		}
	}
	return fieldMap
}

func getType(val interface{}) reflect.Type {
	switch v := val.(type) {
	case reflect.Type:
		return v
	default:
		return reflect.TypeOf(val)
	}
}
