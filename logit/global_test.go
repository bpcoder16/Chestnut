package logit

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/bpcoder16/Chestnut/v4/core/log"
)

func TestGlobalLog(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := log.NewStdLogger(buf)
	SetLogger(logger)

	testCases := []struct {
		level   log.Level
		content []interface{}
	}{
		{
			log.LevelDebug,
			[]interface{}{"test debug"},
		},
		{
			log.LevelInfo,
			[]interface{}{"test info"},
		},
		{
			log.LevelInfo,
			[]interface{}{"test %s", "info"},
		},
		{
			log.LevelWarn,
			[]interface{}{"test warn"},
		},
		{
			log.LevelError,
			[]interface{}{"test error"},
		},
		{
			log.LevelError,
			[]interface{}{"test %s", "error"},
		},
	}

	var expected []string
	for _, testCase := range testCases {
		msg := fmt.Sprintf(testCase.content[0].(string), testCase.content[1:]...)
		switch testCase.level {
		case log.LevelDebug:
			Debug(msg)
			expected = append(expected, fmt.Sprintf("%s %s=%s", "DEBUG", log.DefaultMessageKey, msg))
			DebugF(testCase.content[0].(string), testCase.content[1:]...)
			expected = append(expected, fmt.Sprintf("%s %s=%s", "DEBUG", log.DefaultMessageKey, msg))
			DebugW("logit", msg)
			expected = append(expected, fmt.Sprintf("%s logit=%s", "DEBUG", msg))
		case log.LevelInfo:
			Info(msg)
			expected = append(expected, fmt.Sprintf("%s %s=%s", "INFO", log.DefaultMessageKey, msg))
			InfoF(testCase.content[0].(string), testCase.content[1:]...)
			expected = append(expected, fmt.Sprintf("%s %s=%s", "INFO", log.DefaultMessageKey, msg))
			InfoW("logit", msg)
			expected = append(expected, fmt.Sprintf("%s logit=%s", "INFO", msg))
		case log.LevelWarn:
			Warn(msg)
			expected = append(expected, fmt.Sprintf("%s %s=%s", "WARN", log.DefaultMessageKey, msg))
			WarnF(testCase.content[0].(string), testCase.content[1:]...)
			expected = append(expected, fmt.Sprintf("%s %s=%s", "WARN", log.DefaultMessageKey, msg))
			WarnW("logit", msg)
			expected = append(expected, fmt.Sprintf("%s logit=%s", "WARN", msg))
		case log.LevelError:
			Error(msg)
			expected = append(expected, fmt.Sprintf("%s %s=%s", "ERROR", log.DefaultMessageKey, msg))
			ErrorF(testCase.content[0].(string), testCase.content[1:]...)
			expected = append(expected, fmt.Sprintf("%s %s=%s", "ERROR", log.DefaultMessageKey, msg))
			ErrorW("logit", msg)
			expected = append(expected, fmt.Sprintf("%s logit=%s", "ERROR", msg))
		default:
		}
	}
	_ = Log(log.LevelInfo, log.DefaultMessageKey, "test logit")
	expected = append(expected, fmt.Sprintf("%s %s=%s", "INFO", log.DefaultMessageKey, "test logit"))

	expected = append(expected, "")

	t.Logf("Content: %s", buf.String())

	if buf.String() != strings.Join(expected, "\n") {
		t.Errorf("Expected: %s, got: %s", strings.Join(expected, "\n"), buf.String())
	}
}

func TestGlobalContext(t *testing.T) {
	buf := new(bytes.Buffer)
	SetLogger(log.NewStdLogger(buf))
	Context(context.Background()).InfoF("111")
	expected := fmt.Sprintf("INFO %s=111\n", log.DefaultMessageKey)
	if buf.String() != expected {
		t.Errorf("Expected:%s, got:%s", expected, buf.String())
	}
}
