package conv

import (
	"strconv"
	"unsafe"
)

func StringToInt(str string) (int, error) {
	return strconv.Atoi(str)
}

func StringToInt32(str string) (int32, error) {
	i, err := strconv.ParseInt(str, 10, 32)
	return int32(i), err
}

func StringToInt64(str string) (int64, error) {
	return strconv.ParseInt(str, 10, 64)
}

func IntToString(int int) string {
	return strconv.Itoa(int)
}

func Int32ToString(int32 int32) string {
	return strconv.FormatInt(int64(int32), 10)
}

func Int64ToString(int64 int64) string {
	return strconv.FormatInt(int64, 10)
}

func StringToFloat64(str string) (float64, error) {
	// float32, _ := strconv.ParseFloat(str, 32)
	// float64, _ := strconv.ParseFloat(str, 64)
	return strconv.ParseFloat(str, 64)
}

func Float64ToString(float64 float64) string {
	return strconv.FormatFloat(float64, 'f', 2, 64)
}

func StringToBytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
