package sugar

import (
	"testing"
	"time"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestTimestampSerialize(t *testing.T) {
	tt := []struct {
		input  interface{}
		output interface{}
	}{
		{
			input:  time.Date(2017, time.June, 23, 12, 23, 34, 45, time.UTC),
			output: "2017-06-23T12:23:34.000000045Z",
		},
		{
			input: pq.NullTime{
				Time:  time.Date(2017, time.June, 23, 12, 23, 34, 45, time.UTC),
				Valid: true,
			},
			output: "2017-06-23T12:23:34.000000045Z",
		},
		{
			input: pq.NullTime{
				Time:  time.Date(2017, time.June, 23, 12, 23, 34, 45, time.UTC),
				Valid: false,
			},
			output: nil,
		},
	}

	for _, tc := range tt {
		assert.Equal(t, tc.output, timeSerialize(tc.input))
	}
}

func TestTimestampParse(t *testing.T) {
	tt := []struct {
		input  interface{}
		output interface{}
	}{
		{
			input:  "2017-06-23T12:23:34.000000045Z",
			output: time.Date(2017, time.June, 23, 12, 23, 34, 45, time.UTC),
		},
		{
			input:  4,
			output: nil,
		},
	}

	for _, tc := range tt {
		result := timeParseValue(tc.input)
		assert.Equal(t, tc.output, result)
	}
}

func TestTimestampParseLiteral(t *testing.T) {
	v := &ast.StringValue{
		Kind:  "foo",
		Loc:   &ast.Location{},
		Value: "2017-06-23T12:23:34.000000045Z",
	}
	res := timeParseLiteral(v)
	assert.Equal(t, time.Date(2017, time.June, 23, 12, 23, 34, 45, time.UTC), res)
}
