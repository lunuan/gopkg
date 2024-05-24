package json

import (
	jsoniter "github.com/json-iterator/go"
)

var jsoniterAPI = jsoniter.ConfigCompatibleWithStandardLibrary

func Marshal(v interface{}) ([]byte, error) {
	return jsoniterAPI.Marshal(v)
}

func Unmarshal(data []byte, v interface{}) error {
	return jsoniterAPI.Unmarshal(data, v)
}

func MarshalToString(v interface{}) (string, error) {
	return jsoniterAPI.MarshalToString(v)
}

func UnmarshalFromString(data string, v interface{}) error {
	return jsoniterAPI.UnmarshalFromString(data, v)
}
