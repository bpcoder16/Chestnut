package log

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

var _ Logger = (*stdLogger)(nil)

type stdLogger struct {
	w         io.Writer
	isDiscard bool
	mu        sync.Mutex
	pool      *sync.Pool
}

func (l *stdLogger) Log(level Level, keyValues ...interface{}) error {
	if l.isDiscard || len(keyValues) == 0 {
		return nil
	}
	if len(keyValues)%2 != 0 {
		keyValues = append(keyValues, "keyValues unpaired")
	}

	// 使用 sync.Pool 对象池，提高并发能力
	buf := l.pool.Get().(*bytes.Buffer)
	defer l.pool.Put(buf)

	buf.WriteString(level.String())
	for i := 0; i < len(keyValues); i += 2 {
		_, _ = fmt.Fprintf(buf, " %s=%v", keyValues[i], keyValues[i+1])
	}
	buf.WriteByte('\n')
	defer buf.Reset()

	l.mu.Lock()
	defer l.mu.Unlock()
	_, err := l.w.Write(buf.Bytes())
	return err
}

func NewStdLogger(w io.Writer) Logger {
	return &stdLogger{
		w:         w,
		isDiscard: w == io.Discard,
		pool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}
