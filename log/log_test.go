package log

import "testing"

func TestKvEncoder(t *testing.T) {

	log := NewSugaredLogger(&Config{
		Format: "common",
		Level:  "debug",
	})

	message := "this a log message"
	log.Debug(message)
	log.Debugf("%s", message)
	log.Debugw(
		message,
		"int", 1,
		"bool", true,
		"string", "str",
		"float", 1.1,
		"complex", 1+1i,
		"duration", 1,
		"byte", []byte("byte"),
		"int64", int64(1),
		"int32", int32(1),
		"int16", int16(1),
		"float32", float32(1.1),
		"float64", float64(1.1),
		"slice", []int{1, 2, 3},
		"map", map[string]string{"k": "v"},
		"array", [3]int{1, 2, 3},
		"interface", interface{}(1),
		"chan", make(chan int),
		"ptr", &message,
		"func", func() {},
		"struct", struct{}{},
	)
}
