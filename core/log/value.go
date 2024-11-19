package log

import (
	"context"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	// DefaultCaller is a Valuer that returns the file and line.
	DefaultCaller = Caller(4)

	// DefaultTimestamp is a Valuer that returns the current wallclock time.
	DefaultTimestamp = Timestamp(time.RFC3339)
)

var zeroSourceDir string

func init() {
	_, file, _, _ := runtime.Caller(0)
	// compatible solution to get gorm source directory with various operating systems
	zeroSourceDir = sourceDir(file)
}

func sourceDir(file string) string {
	dir := filepath.Dir(file)
	dir = filepath.Dir(dir)

	s := filepath.Dir(dir)
	if filepath.Base(s) != "zero" {
		s = dir
	}
	return filepath.ToSlash(s) + "/"
}

// Valuer is returns a logit value.
type Valuer func(ctx context.Context) interface{}

// Value return the function value.
func Value(ctx context.Context, v interface{}) interface{} {
	if v, ok := v.(Valuer); ok {
		return v(ctx)
	}
	return v
}

// Caller returns a Valuer that returns a pkg/file:line description of the caller.
func Caller(depth int) Valuer {
	return func(_ context.Context) interface{} {
		_, file, line, _ := runtime.Caller(depth)
		idx := strings.LastIndexByte(file, '/')
		if idx == -1 {
			return file[idx+1:] + ":" + strconv.Itoa(line)
		}
		idx = strings.LastIndexByte(file[:idx], '/')
		return file[idx+1:] + ":" + strconv.Itoa(line)
	}
}

func FileWithLineNumCaller() Valuer {
	return func(_ context.Context) interface{} {
		pcs := [13]uintptr{}
		// the third caller usually from gorm internal
		length := runtime.Callers(3, pcs[:])
		frames := runtime.CallersFrames(pcs[:length])
		for i := 0; i < length; i++ {
			// second return value is "more", not "ok"
			frame, _ := frames.Next()
			if (!strings.HasPrefix(frame.File, zeroSourceDir) ||
				strings.HasSuffix(frame.File, "_test.go")) && !strings.HasSuffix(frame.File, ".gen.go") {
				return string(strconv.AppendInt(append([]byte(frame.File), ':'), int64(frame.Line), 10))
			}
		}

		return ""
	}
}

func FileWithLineNumCallerRedis() Valuer {
	return func(_ context.Context) interface{} {
		pcs := [13]uintptr{}
		// the third caller usually from gorm internal
		length := runtime.Callers(3, pcs[:])
		frames := runtime.CallersFrames(pcs[:length])
		for i := 0; i < length; i++ {
			// second return value is "more", not "ok"
			frame, _ := frames.Next()
			// TODO 很尴尬的处理方式，等我后续有想法再改吧
			if !strings.HasPrefix(frame.File, zeroSourceDir) && !strings.Contains(frame.File, "github.com/redis/go-redis") && !strings.HasSuffix(frame.File, ".gen.go") {
				return string(strconv.AppendInt(append([]byte(frame.File), ':'), int64(frame.Line), 10))
			}
		}

		return ""
	}
}

// Timestamp returns a timestamp Valuer with a custom time format.
func Timestamp(layout string) Valuer {
	return func(_ context.Context) interface{} {
		return time.Now().Format(layout)
	}
}

func bindValues(ctx context.Context, keyValues []interface{}) {
	for i := 1; i < len(keyValues); i += 2 {
		if v, ok := keyValues[i].(Valuer); ok {
			keyValues[i] = v(ctx)
		}
	}
}

func containsValuer(keyValues []interface{}) bool {
	for i := 1; i < len(keyValues); i += 2 {
		if _, ok := keyValues[i].(Valuer); ok {
			return true
		}
	}
	return false
}
