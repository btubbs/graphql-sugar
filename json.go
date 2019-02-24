package sugar

import (
	"encoding/json"
	"strconv"

	"github.com/btubbs/pqjson"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

// JSON is a custom scalar type for using custom json blobs as leaves on our responses.
var JSON = graphql.NewScalar(
	graphql.ScalarConfig{
		Name:         "JSON",
		Description:  "JSON is a custom scalar for attaching json blobs as leaves on a graphql API request or response.",
		Serialize:    JSONSerialize,
		ParseValue:   JSONParseValue,
		ParseLiteral: JSONParseLiteral,
	})

// JSONSerialize serializes objects into the custom JSON blob graphql scalar type.
func JSONSerialize(value interface{}) interface{} {
	switch value := value.(type) {
	case pqjson.RawMessage:
		return json.RawMessage(value)
	default:
		return nil
	}
}

// JSONParseValue implements the  ParseValue so that JSON can satisfy the graphql.Scalar interface.
// This method gets called when parsing JSON type obejcts out of query variables
// This is also called when doing the query: variables: pattern.  Assume
// that this can be Marshalled like below
func JSONParseValue(value interface{}) interface{} {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return json.RawMessage(raw)
}

// JSONParseLiteral converts fragments of a graphql AST into a json Raw Message. It's used for parsing input
// objects for our custom scalar JSON type. The assumption is that the graphql query object is close enough
// to json that we can convert it.
func JSONParseLiteral(valueAST ast.Value) interface{} {
	parsedAST := parseAST(valueAST)
	bs, err := json.Marshal(parsedAST)
	if err != nil {
		return nil
	}

	return json.RawMessage(bs)
}

func parseAST(astValue ast.Value) interface{} {
	switch astValue := astValue.(type) {
	case *ast.BooleanValue, *ast.StringValue:
		return astValue.GetValue()
	case *ast.FloatValue, *ast.IntValue:
		val, err := strconv.ParseFloat(astValue.GetValue().(string), 0)
		if err != nil {
			return nil
		}
		return val
	case *ast.ObjectField:
		res := map[string]interface{}{}
		k := astValue.Name.Value
		v := parseAST(astValue.Value)
		res[k] = v
		return res
	case *ast.ObjectValue:
		res := map[string]interface{}{}
		for _, field := range astValue.Fields {
			fieldMap := parseAST(field)
			for k, v := range fieldMap.(map[string]interface{}) {
				res[k] = v
			}
		}
		return res
	case *ast.ListValue:
		res := []interface{}{}
		for _, val := range astValue.Values {
			res = append(res, parseAST(val))
		}
		return res
	default:
		return nil
	}
}
