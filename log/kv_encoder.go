package log

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"os"

	// "encoding/json"
	"fmt"
	"math"
	"time"
	"unicode/utf8"

	"github.com/lunuan/gopkg/log/bufferpool"
	"github.com/lunuan/gopkg/log/pool"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

const (
	// For JSON-escaping; see jsonEncoder.safeAddString below.
	_hex = "0123456789abcdef"
)

var (
	_jsonPool        = pool.New(func() *kvEncoder { return &kvEncoder{} })
	nullLiteralBytes = []byte("null")
	encodeErrorBytes = []byte("encodeError")
)

type kvEncoder struct {
	*zapcore.EncoderConfig
	buf            *buffer.Buffer
	openNamespaces int
	hostname       string

	// for encoding generic values by reflection
	reflectBuf *buffer.Buffer
	reflectEnc zapcore.ReflectedEncoder
}

func NewkvEncoder(cfg zapcore.EncoderConfig) *kvEncoder {
	// spaced := false
	if cfg.SkipLineEnding {
		cfg.LineEnding = ""
	} else if cfg.LineEnding == "" {
		cfg.LineEnding = "\n"
	}
	// If no EncoderConfig.NewReflectedEncoder is provided by the user, then use default
	if cfg.NewReflectedEncoder == nil {
		cfg.NewReflectedEncoder = defaultReflectedEncoder
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	return &kvEncoder{
		EncoderConfig: &cfg,
		buf:           bufferpool.Get(),
		hostname:      hostname,
	}
}

func (enc kvEncoder) Clone() zapcore.Encoder {
	clone := _jsonPool.Get()
	clone.EncoderConfig = enc.EncoderConfig
	// clone.spaced = enc.spaced
	clone.openNamespaces = enc.openNamespaces
	clone.buf = bufferpool.Get()
	clone.buf.Write(enc.buf.Bytes())
	return clone
}

func (c kvEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	line := bufferpool.Get()

	// We don't want the entry's metadata to be quoted and escaped (if it's
	// encoded as strings), which means that we can't use the JSON encoder. The
	// simplest option is to use the memory encoder and fmt.Fprint.
	//
	// If this ever becomes a performance bottleneck, we can implement
	// ArrayEncoder for our plain-text format.
	arr := getSliceEncoder()
	if c.TimeKey != "" && c.EncodeTime != nil && !ent.Time.IsZero() {
		c.EncodeTime(ent.Time, arr)
	}
	if c.LevelKey != "" && c.EncodeLevel != nil {
		c.EncodeLevel(ent.Level, arr)
	}

	arr.AppendString(c.hostname)

	if ent.LoggerName != "" && c.NameKey != "" {
		nameEncoder := c.EncodeName

		if nameEncoder == nil {
			// Fall back to FullNameEncoder for backward compatibility.
			nameEncoder = zapcore.FullNameEncoder
		}

		nameEncoder(ent.LoggerName, arr)
	}
	if ent.Caller.Defined {
		if c.CallerKey != "" && c.EncodeCaller != nil {
			c.EncodeCaller(ent.Caller, arr)
		}
		if c.FunctionKey != "" {
			arr.AppendString(ent.Caller.Function)
		}
	}
	for i := range arr.elems {
		if i > 0 {
			line.AppendString(c.ConsoleSeparator)
		}
		fmt.Fprint(line, arr.elems[i])
	}
	putSliceEncoder(arr)

	// Add any structured context.
	c.writeContext(line, fields)

	// Add the message itself.
	if c.MessageKey != "" {
		c.addSeparatorIfNecessary(line)
		line.AppendString(ent.Message)
	}

	// If there's no stacktrace key, honor that; this allows users to force
	// single-line output.
	if ent.Stack != "" && c.StacktraceKey != "" {
		line.AppendByte('\n')
		line.AppendString(ent.Stack)
	}

	line.AppendString(c.LineEnding)
	return line, nil
}

func (enc *kvEncoder) OpenNamespace(key string) {
	enc.addKey(key)
	enc.buf.AppendByte('{')
	enc.openNamespaces++
}

func (enc *kvEncoder) closeOpenNamespaces() {
	for i := 0; i < enc.openNamespaces; i++ {
		enc.buf.AppendByte('}')
	}
	enc.openNamespaces = 0
}

func (c kvEncoder) writeContext(line *buffer.Buffer, extra []zapcore.Field) {
	context := c.Clone().(*kvEncoder)
	defer func() {
		// putJSONEncoder assumes the buffer is still used, but we write out the buffer so
		// we can free it.
		context.buf.Free()

		if context.reflectBuf != nil {
			context.reflectBuf.Free()
		}
		context.EncoderConfig = nil
		context.buf = nil
		context.openNamespaces = 0
		context.reflectBuf = nil
		context.reflectEnc = nil
		_jsonPool.Put(context)
	}()

	for i := range extra {
		// fmt.Println(extra[i].Type, extra[i].Key)
		extra[i].AddTo(context)
	}

	context.closeOpenNamespaces()
	if context.buf.Len() == 0 {
		return
	}

	c.addSeparatorIfNecessary(line)

	line.Write(context.buf.Bytes())

}

func (c kvEncoder) addSeparatorIfNecessary(line *buffer.Buffer) {
	if line.Len() > 0 {
		line.AppendString(c.ConsoleSeparator)
	}
}

// safeAppendStringLike is a generic implementation of safeAddString and safeAddByteString.
// It appends a string or byte slice to the buffer, escaping all special characters.
func safeAppendStringLike[S []byte | string](
	// appendTo appends this string-like object to the buffer.
	appendTo func(*buffer.Buffer, S),
	// decodeRune decodes the next rune from the string-like object
	// and returns its value and width in bytes.
	decodeRune func(S) (rune, int),
	buf *buffer.Buffer,
	s S,
) {
	// The encoding logic below works by skipping over characters
	// that can be safely copied as-is,
	// until a character is found that needs special handling.
	// At that point, we copy everything we've seen so far,
	// and then handle that special character.
	//
	// last is the index of the last byte that was copied to the buffer.
	last := 0
	for i := 0; i < len(s); {
		if s[i] >= utf8.RuneSelf {
			// Character >= RuneSelf may be part of a multi-byte rune.
			// They need to be decoded before we can decide how to handle them.
			r, size := decodeRune(s[i:])
			if r != utf8.RuneError || size != 1 {
				// No special handling required.
				// Skip over this rune and continue.
				i += size
				continue
			}

			// Invalid UTF-8 sequence.
			// Replace it with the Unicode replacement character.
			appendTo(buf, s[last:i])
			buf.AppendString(`\ufffd`)

			i++
			last = i
		} else {
			// Character < RuneSelf is a single-byte UTF-8 rune.
			if s[i] >= 0x20 && s[i] != '\\' && s[i] != '"' {
				// No escaping necessary.
				// Skip over this character and continue.
				i++
				continue
			}

			// This character needs to be escaped.
			appendTo(buf, s[last:i])
			switch s[i] {
			case '\\', '"':
				buf.AppendByte('\\')
				buf.AppendByte(s[i])
			case '\n':
				buf.AppendByte('\\')
				buf.AppendByte('n')
			case '\r':
				buf.AppendByte('\\')
				buf.AppendByte('r')
			case '\t':
				buf.AppendByte('\\')
				buf.AppendByte('t')
			default:
				// Encode bytes < 0x20, except for the escape sequences above.
				buf.AppendString(`\u00`)
				buf.AppendByte(_hex[s[i]>>4])
				buf.AppendByte(_hex[s[i]&0xF])
			}

			i++
			last = i
		}
	}

	// add remaining
	appendTo(buf, s[last:])
}

// safeAddString JSON-escapes a string and appends it to the internal buffer.
// Unlike the standard library's encoder, it doesn't attempt to protect the
// user from browser vulnerabilities or JSONP-related problems.
func (enc *kvEncoder) safeAddString(s string) {
	safeAppendStringLike(
		(*buffer.Buffer).AppendString,
		utf8.DecodeRuneInString,
		enc.buf,
		s,
	)
}

// safeAddByteString is no-alloc equivalent of safeAddString(string(s)) for s []byte.
func (enc *kvEncoder) safeAddByteString(s []byte) {
	safeAppendStringLike(
		(*buffer.Buffer).AppendBytes,
		utf8.DecodeRune,
		enc.buf,
		s,
	)
}

func (enc *kvEncoder) truncate() {
	enc.buf.Reset()
}

func (enc *kvEncoder) addKey(key string) {
	enc.addElementSeparator()
	enc.safeAddString(key)
	enc.buf.AppendByte('=')
}

func (enc *kvEncoder) addElementSeparator() {
	last := enc.buf.Len() - 1
	if last < 0 {
		return
	}
	switch enc.buf.Bytes()[last] {
	case '{', '[', ':', ',', ' ':
		return
	default:
		enc.buf.AppendByte(' ')
	}
}

func (enc *kvEncoder) AddInt(k string, v int)               { enc.AddInt64(k, int64(v)) }
func (enc *kvEncoder) AddInt8(k string, v int8)             { enc.AddInt64(k, int64(v)) }
func (enc *kvEncoder) AddInt16(k string, v int16)           { enc.AddInt64(k, int64(v)) }
func (enc *kvEncoder) AddInt32(k string, v int32)           { enc.AddInt64(k, int64(v)) }
func (enc *kvEncoder) AddInt64(key string, val int64)       { enc.addKey(key); enc.AppendInt64(val) }
func (enc *kvEncoder) AddUintptr(k string, v uintptr)       { enc.AddUint64(k, uint64(v)) }
func (enc *kvEncoder) AddUint(k string, v uint)             { enc.AddUint64(k, uint64(v)) }
func (enc *kvEncoder) AddUint8(k string, v uint8)           { enc.AddUint64(k, uint64(v)) }
func (enc *kvEncoder) AddUint16(k string, v uint16)         { enc.AddUint64(k, uint64(v)) }
func (enc *kvEncoder) AddUint32(k string, v uint32)         { enc.AddUint64(k, uint64(v)) }
func (enc *kvEncoder) AddUint64(key string, val uint64)     { enc.addKey(key); enc.AppendUint64(val) }
func (enc *kvEncoder) AddFloat32(key string, val float32)   { enc.addKey(key); enc.AppendFloat32(val) }
func (enc *kvEncoder) AddFloat64(key string, val float64)   { enc.addKey(key); enc.AppendFloat64(val) }
func (enc *kvEncoder) AddBool(key string, val bool)         { enc.addKey(key); enc.AppendBool(val) }
func (enc *kvEncoder) AddString(key, val string)            { enc.addKey(key); enc.AppendString(val) }
func (enc *kvEncoder) AddTime(key string, val time.Time)    { enc.addKey(key); enc.AppendTime(val) }
func (enc *kvEncoder) AddComplex64(k string, v complex64)   { enc.addKey(k); enc.AppendComplex64(v) }
func (enc *kvEncoder) AddComplex128(k string, v complex128) { enc.addKey(k); enc.AppendComplex128(v) }
func (enc *kvEncoder) AppendComplex64(v complex64)          { enc.appendComplex(complex128(v), 32) }
func (enc *kvEncoder) AppendComplex128(v complex128)        { enc.appendComplex(complex128(v), 64) }
func (enc *kvEncoder) AppendFloat64(v float64)              { enc.appendFloat(v, 64) }
func (enc *kvEncoder) AppendFloat32(v float32)              { enc.appendFloat(float64(v), 32) }
func (enc *kvEncoder) AppendInt(v int)                      { enc.AppendInt64(int64(v)) }
func (enc *kvEncoder) AppendInt32(v int32)                  { enc.AppendInt64(int64(v)) }
func (enc *kvEncoder) AppendInt16(v int16)                  { enc.AppendInt64(int64(v)) }
func (enc *kvEncoder) AppendInt8(v int8)                    { enc.AppendInt64(int64(v)) }
func (enc *kvEncoder) AppendUint(v uint)                    { enc.AppendUint64(uint64(v)) }
func (enc *kvEncoder) AppendUint32(v uint32)                { enc.AppendUint64(uint64(v)) }
func (enc *kvEncoder) AppendUint16(v uint16)                { enc.AppendUint64(uint64(v)) }
func (enc *kvEncoder) AppendUint8(v uint8)                  { enc.AppendUint64(uint64(v)) }
func (enc *kvEncoder) AppendUintptr(v uintptr)              { enc.AppendUint64(uint64(v)) }
func (enc *kvEncoder) AppendUint64(val uint64)              { enc.buf.AppendUint(val) }
func (enc *kvEncoder) AppendBool(val bool)                  { enc.buf.AppendBool(val) }
func (enc *kvEncoder) AppendInt64(val int64)                { enc.buf.AppendInt(val) }

func (enc *kvEncoder) appendFloat(val float64, bitSize int) {
	switch {
	case math.IsNaN(val):
		enc.buf.AppendString(`"NaN"`)
	case math.IsInf(val, 1):
		enc.buf.AppendString(`"+Inf"`)
	case math.IsInf(val, -1):
		enc.buf.AppendString(`"-Inf"`)
	default:
		enc.buf.AppendFloat(val, bitSize)
	}
}

// appendComplex appends the encoded form of the provided complex128 value.
// precision specifies the encoding precision for the real and imaginary
// components of the complex number.
func (enc *kvEncoder) appendComplex(val complex128, precision int) {
	// enc.addElementSeparator()
	// Cast to a platform-independent, fixed-size type.
	r, i := float64(real(val)), float64(imag(val))
	// enc.buf.AppendByte('"')
	// Because we're always in a quoted string, we can use strconv without
	// special-casing NaN and +/-Inf.
	enc.buf.AppendFloat(r, precision)
	// If imaginary part is less than 0, minus (-) sign is added by default
	// by AppendFloat.
	if i >= 0 {
		enc.buf.AppendByte('+')
	}
	enc.buf.AppendFloat(i, precision)
	enc.buf.AppendByte('i')
	// enc.buf.AppendByte('"')
}

func (enc *kvEncoder) AddArray(key string, arr zapcore.ArrayMarshaler) error {
	enc.addKey(key)
	return enc.AppendArray(arr)
}

func (enc *kvEncoder) AppendArray(arr zapcore.ArrayMarshaler) error {
	// enc.addElementSeparator()
	// 	enc.buf.AppendByte('[')
	// 	err := arr.MarshalLogArray(enc)
	// 	enc.buf.AppendByte(']')
	// 	return err
	return enc.AppendReflected(arr)
}

func (enc *kvEncoder) AddBinary(key string, val []byte) {
	enc.AddString(key, base64.StdEncoding.EncodeToString(val))
}

func (enc *kvEncoder) AddByteString(key string, val []byte) {
	enc.addKey(key)
	enc.AppendByteString(val)
}

func (enc *kvEncoder) AppendByteString(val []byte) {
	enc.addElementSeparator()
	enc.buf.AppendByte('"')
	enc.safeAddByteString(val)
	enc.buf.AppendByte('"')
}

func (enc *kvEncoder) AddDuration(key string, val time.Duration) {
	enc.addKey(key)
	enc.AppendDuration(val)
}

func (enc *kvEncoder) AppendDuration(val time.Duration) {
	cur := enc.buf.Len()
	if e := enc.EncodeDuration; e != nil {
		e(val, enc)
	}
	if cur == enc.buf.Len() {
		// User-supplied EncodeDuration is a no-op. Fall back to nanoseconds to keep JSON valid.
		enc.AppendInt64(int64(val))
	}
}

func (enc *kvEncoder) AppendString(val string) {
	// enc.buf.AppendByte('"')
	enc.safeAddString(val)
	// enc.buf.AppendByte('"')
}

func (enc *kvEncoder) AppendTime(val time.Time) {
	cur := enc.buf.Len()
	if e := enc.EncodeTime; e != nil {
		e(val, enc)
	}
	if cur == enc.buf.Len() {
		// User-supplied EncodeTime is a no-op. Fall back to nanos since epoch to keep output JSON valid.
		enc.AppendInt64(val.UnixNano())
	}
}

func (enc *kvEncoder) AddObject(key string, obj zapcore.ObjectMarshaler) error {
	enc.addKey(key)
	return enc.AppendObject(obj)
}

func (enc *kvEncoder) AppendObject(obj zapcore.ObjectMarshaler) error {
	// Close ONLY new openNamespaces that are created during AppendObject().
	old := enc.openNamespaces
	enc.openNamespaces = 0
	enc.addElementSeparator()
	enc.buf.AppendByte('{')
	err := obj.MarshalLogObject(enc)
	enc.buf.AppendByte('}')
	enc.closeOpenNamespaces()
	enc.openNamespaces = old
	return err
}

func defaultReflectedEncoder(w io.Writer) zapcore.ReflectedEncoder {
	enc := json.NewEncoder(w)
	// For consistency with our custom JSON encoder.
	enc.SetEscapeHTML(false)
	return enc
}

func (enc *kvEncoder) resetReflectBuf() {
	if enc.reflectBuf == nil {
		enc.reflectBuf = bufferpool.Get()
		enc.reflectEnc = enc.NewReflectedEncoder(enc.reflectBuf)
	} else {
		enc.reflectBuf.Reset()
	}
}

func (enc *kvEncoder) encodeReflected(obj interface{}) ([]byte, error) {
	if obj == nil {
		return nullLiteralBytes, nil
	}
	enc.resetReflectBuf()
	if err := enc.reflectEnc.Encode(obj); err != nil {
		// return nil, err
		return encodeErrorBytes, nil
	}
	enc.reflectBuf.TrimNewline()
	return enc.reflectBuf.Bytes(), nil
}

func (enc *kvEncoder) AddReflected(key string, obj interface{}) error {
	valueBytes, err := enc.encodeReflected(obj)
	if err != nil {
		return err
	}
	enc.addKey(key)
	_, err = enc.buf.Write(valueBytes)
	return err
}

func (enc *kvEncoder) AppendReflected(val interface{}) error {
	valueBytes, err := enc.encodeReflected(val)
	if err != nil {
		return err
	}
	_, err = enc.buf.Write(valueBytes)
	return err
}
