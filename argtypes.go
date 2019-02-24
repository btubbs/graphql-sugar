package sugar

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/btubbs/datetime"
	"github.com/btubbs/pqjson"
	"github.com/guregu/null"
)

// LoadUInt loads `uint` from graphql arg
func LoadUInt(i interface{}) (uint, error) {
	stringified := fmt.Sprintf("%v", i)
	wd, err := strconv.ParseUint(stringified, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%v is not an unsigned integer", i)
	}

	return uint(wd), nil
}

// LoadBool loads `bool` from graphql arg
func LoadBool(i interface{}) (bool, error) {
	b, ok := i.(bool)
	if !ok {
		return false, fmt.Errorf("%v is not a bool", i)
	}
	return b, nil
}

// LoadBoolPointer loads `*bool` from graphql arg
func LoadBoolPointer(i interface{}) (*bool, error) {
	if i == nil {
		return nil, nil
	}
	b, err := LoadBool(i)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// LoadString loads `string` from graphql arg
func LoadString(i interface{}) (string, error) {
	b, ok := i.(string)
	if !ok {
		return "", fmt.Errorf("%v is not a string", i)
	}
	return b, nil
}

// LoadInt loads `int` from graphql arg
func LoadInt(i interface{}) (int, error) {
	b, ok := i.(int)
	if !ok {
		return 0, fmt.Errorf("%v is not an int", i)
	}
	return b, nil
}

// LoadNullInt loads `null.Int` from graphql arg
func LoadNullInt(i interface{}) (null.Int, error) {
	b, ok := i.(int)
	if !ok {
		return null.Int{}, fmt.Errorf("%v is not an int", i)
	}
	return null.IntFrom(int64(b)), nil
}

// LoadNullString will attempt to load a `null.String` from graphql arg
func LoadNullString(i interface{}) (null.String, error) {
	str, ok := i.(string)
	if !ok {
		return null.String{}, fmt.Errorf("%v is not a string", i)
	}
	return null.StringFrom(str), nil
}

// LoadFloat loads `float` from graphql arg
func LoadFloat(i interface{}) (float64, error) {
	b, ok := i.(float64)
	if !ok {
		return 0, fmt.Errorf("%v is not a float", i)
	}
	return b, nil
}

// LoadTime loads `time.Time` from graphql arg
func LoadTime(i interface{}) (time.Time, error) {
	switch s := i.(type) {
	case string:
		return datetime.Parse(s, time.UTC)
	case time.Time:
		return s, nil
	}

	return time.Time{}, fmt.Errorf("%v is not a ISO8601 timestamp", i)
}

// LoadRawJSON loads `pqjson.RawMessage` from graphql arg
func LoadRawJSON(i interface{}) (pqjson.RawMessage, error) {
	switch value := i.(type) {
	case json.RawMessage:
		return pqjson.RawMessage(value), nil
	default:
		return pqjson.RawMessage(""), fmt.Errorf("%v is not a json.RawMessage", i)
	}
}
