package log

import (

	// "encoding/json"
	"fmt"

	"github.com/lunuan/gopkg/log/bufferpool"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

type CommonEncoder struct {
	*kvEncoder
}

func NewCommonEncoder(cfg zapcore.EncoderConfig) *CommonEncoder {
	return &CommonEncoder{NewkvEncoder(cfg)}
}

func (c CommonEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	line := bufferpool.Get()

	// We don't want the entry's metadata to be quoted and escaped (if it's
	// encoded as strings), which means that we can't use the JSON encoder. The
	// simplest option is to use the memory encoder and fmt.Fprint.
	//
	// If this ever becomes a performance bottleneck, we can implement
	// ArrayEncoder for our plain-text format.
	arr := getSliceEncoder()
	// timestrap
	if c.TimeKey != "" && c.EncodeTime != nil && !ent.Time.IsZero() {
		c.EncodeTime(ent.Time, arr)
	}
	// [process]
	arr.AppendString("[main]")
	// hostname
	arr.AppendString(c.hostname)
	// level
	if c.LevelKey != "" && c.EncodeLevel != nil {
		c.EncodeLevel(ent.Level, arr)
	}

	// caller
	if ent.Caller.Defined && c.EncodeCaller != nil {
		c.EncodeCaller(ent.Caller, arr)
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
