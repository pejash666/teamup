package util

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
)

func BoolToDB(anyBool bool) int {
	if anyBool {
		return 1
	} else {
		return 0
	}
}

func ToReadable(value interface{}) string {
	if value == nil {
		return "nil"
	}
	var str string
	switch vt := value.(type) {
	case string:
		str = vt
	case []byte:
		str = string(vt)
	default:
		kind := reflect.TypeOf(value).Kind()
		switch kind {
		case reflect.Struct, reflect.Interface, reflect.Map, reflect.Slice:
			bytes, err := json.MarshalIndent(value, "", " ")
			if err != nil {
				return fmt.Sprintf("format error: %v", err)
			}
			str = string(bytes)
		default:
			str = fmt.Sprint(value)
		}
	}
	return str
}

func Round(f float64, n int) float64 {
	pow10_n := math.Pow10(n)
	return math.Trunc((f+0.5/pow10_n)*pow10_n) / pow10_n
}
