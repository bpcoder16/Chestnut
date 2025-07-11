package gin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bpcoder16/Chestnut/v2/appconfig/env"
	"net/http"
	"os"
	"runtime"
	"strings"
)

const (
	fuzzyStr = "***"
)

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
	reset     = string([]byte{27, 91, 48, 109})
)

var needFilterMap = map[string]struct{}{
	"password": {},
	"token":    {},
}

func filterHeader(header http.Header) http.Header {
	if env.RunMode() != env.RunModeRelease {
		return header
	}
	for key := range header {
		if _, ok := needFilterMap[strings.ToLower(key)]; ok {
			header.Set(key, fuzzyStr)
		}
	}
	return header
}

func filterBody(b []byte) interface{} {
	var reqBody map[string]interface{}
	err := json.Unmarshal(b, &reqBody)
	if err != nil {
		return string(b)
	}
	if env.RunMode() != env.RunModeRelease {
		return reqBody
	}
	for key := range reqBody {
		if _, ok := needFilterMap[strings.ToLower(key)]; ok {
			reqBody[key] = fuzzyStr
		}
	}
	return reqBody
}

// stack returns a nicely formatted stack frame, skipping skip frames.
func stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		_, _ = fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		_, _ = fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	if lastSlash := bytes.LastIndex(name, slash); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}

func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}
