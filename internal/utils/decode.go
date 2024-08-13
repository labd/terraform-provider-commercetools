package utils

import (
	"github.com/mitchellh/mapstructure"
	"reflect"
	"time"
)

func toTimeHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		switch f.Kind() {
		case reflect.String:
			return time.Parse(time.RFC3339, data.(string))
		case reflect.Float64:
			return time.Unix(0, int64(data.(float64))*int64(time.Millisecond)), nil
		case reflect.Int64:
			return time.Unix(0, data.(int64)*int64(time.Millisecond)), nil
		default:
			return data, nil
		}
		// Convert it by parsing
	}
}

func DecodeStruct(input interface{}, result interface{}) error {
	meta := &mapstructure.Metadata{}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: meta,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			toTimeHookFunc()),
		Result: result,
	})
	if err != nil {
		return err
	}

	if err := decoder.Decode(input); err != nil {
		return err
	}

	if val, ok := result.(Decoder); ok {
		if raw, ok := input.(map[string]interface{}); ok {
			unused := make(map[string]interface{})
			for _, key := range meta.Unused {
				unused[key] = raw[key]
			}
			val.DecodeStruct(unused)
		}
	}

	return err
}

type Decoder interface {
	DecodeStruct(map[string]interface{}) error
}
