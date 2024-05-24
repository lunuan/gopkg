package controller

type Controller interface {
	GetInt(k string, def int) int
	GetInt32(k string, def int32) int32
	GetInt64(k string, def int64) int64
	GetString(k string, def string) string
	GetFloat64(k string, def float64) float64
	GetBool(k string, def bool) bool
	ShouldBind(v interface{}) error
}
