package sugar

import (
	"time"

	"github.com/btubbs/datetime"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/guregu/null"
	"github.com/lib/pq"
)

func timeSerialize(value interface{}) interface{} {
	switch value := value.(type) {
	case time.Time:
		buff, err := value.MarshalText()
		if err != nil {
			return nil
		}

		return string(buff)
	case *time.Time:
		return timeSerialize(*value)
	case pq.NullTime:
		if value.Valid {
			return timeSerialize(value.Time)
		}
		return nil
	case null.Time:
		if value.Valid {
			return timeSerialize(value.Time)
		}
		return nil
	default:
		return nil
	}
}

func timeParseValue(value interface{}) interface{} {
	switch value := value.(type) {
	case []byte:
		t, err := datetime.Parse(string(value), time.UTC)
		if err != nil {
			return nil
		}

		return t
	case string:
		return timeParseValue([]byte(value))
	case *string:
		return timeParseValue([]byte(*value))
	default:
		return nil
	}
}

func timeParseLiteral(valueAST ast.Value) interface{} {
	switch valueAST := valueAST.(type) {
	case *ast.StringValue:
		t, err := datetime.Parse(string(valueAST.Value), time.UTC)
		if err != nil {
			return nil
		}

		return t
	}
	return nil
}

// Timestamp is our custom scalar for handling datetimes.
var Timestamp = graphql.NewScalar(graphql.ScalarConfig{
	Name:         "Timestamp",
	Description:  "Timestamp is an ISO8601-formatted date/time string. Values that omit a time zone are assumed to be UTC.",
	Serialize:    timeSerialize,
	ParseValue:   timeParseValue,
	ParseLiteral: timeParseLiteral,
})
