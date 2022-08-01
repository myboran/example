package mjson

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func escapeString(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, "\n", "\\n"), "\r", "\\r"), "\"", "\\\"")
}

func skipParsing(kind reflect.Kind) bool {
	switch kind {
	case reflect.Chan, reflect.Complex64, reflect.Complex128, reflect.Func, reflect.Invalid:
		return true
	default:
		return false
	}
}

func Marshal(v interface{}) (string, error) {
	if v == nil {
		return "null", nil
	}
	t := reflect.TypeOf(v)
	kind := t.Kind()
	if skipParsing(kind) {
		return "", nil
	}
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8,
		reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Bool:
		return fmt.Sprintf("%v", v), nil
	case reflect.String:
		return "\"" + escapeString(v.(string)) + "\"", nil
	case reflect.Array, reflect.Slice:
		vArray := reflect.ValueOf(v)
		results := make([]string, vArray.Len())
		for i := 0; i < vArray.Len(); i++ {
			result, err := Marshal(vArray.Index(i).Interface())
			if err != nil {
				return "", err
			}
			results[i] = result
		}
		return fmt.Sprintf("[%v]", strings.Join(results, ",")), nil
	case reflect.Map:
		vMap := reflect.ValueOf(v)
		keys := vMap.MapKeys()
		var results []string
		j := 0
		for _, key := range keys {
			val := vMap.MapIndex(key)
			result := "null"
			if val.IsValid() {
				if !val.CanInterface() {
					j++
					continue
				}
				r, err := Marshal(val.Interface())
				if err != nil {
					return "", err
				}
				result = r
			}
			results = append(results, "\""+escapeString(fmt.Sprintf("%v", key))+"\":"+result)
		}
		return fmt.Sprintf("{%v}", strings.Join(results, ",")), nil
	case reflect.Ptr, reflect.Interface:
		val := reflect.ValueOf(v).Elem()
		if !val.IsValid() {
			return "null", nil
		}
		return Marshal(val.Interface())
	case reflect.Struct:
		value := reflect.ValueOf(v)
		var results []string
		for i := 0; i < value.NumField(); i++ {
			vField := value.Field(i)
			tField := value.Type().Field(i)

			if !vField.CanInterface() || skipParsing(tField.Type.Kind()) {
				continue
			}

			key := tField.Name
			tag := tField.Tag.Get("json")
			if tag != "" {
				key = tag
			}

			result := "null"
			val := reflect.ValueOf(vField.Interface())
			if val.IsValid() {
				r, err := Marshal(val.Interface())
				if err != nil {
					return "", err
				}
				result = r
			}
			results = append(results, "\""+escapeString(key)+"\":"+result)
		}
		return fmt.Sprintf("{%v}", strings.Join(results, ",")), nil
	default:
		return "marshal fail: " + fmt.Sprintf("%v", v), errors.New("marshal fail")
	}

}

func UnMarshal(s string, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("json: Unmarshal (non-pointer or nil)")
	}
	// TODO
	return nil
}
