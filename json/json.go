package json

import (
	jsoniter "github.com/json-iterator/go"
)

var Json = jsoniter.ConfigCompatibleWithStandardLibrary

func Marshal(v interface{}) ([]byte, error) {
	return Json.Marshal(v)
}

func Unmarshal(data []byte, v interface{}) error {
	return Json.Unmarshal(data, v)
}

func MarshalToString(v interface{}) (string, error) {
	return Json.MarshalToString(v)
}

func UnmarshalFromString(str string, v interface{}) error {
	return Json.UnmarshalFromString(str, v)
}
