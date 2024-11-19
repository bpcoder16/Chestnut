package log

import "testing"

func TestLevelKey(t *testing.T) {
	if LevelInfo.Key() != LevelKey {
		t.Errorf("want: %s, got: %s", LevelKey, LevelInfo.Key())
	}
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		name  string
		level Level
		want  string
	}{
		{name: "DEBUG", level: LevelDebug, want: "DEBUG"},
		{name: "INFO", level: LevelInfo, want: "INFO"},
		{name: "WARN", level: LevelWarn, want: "WARN"},
		{name: "ERROR", level: LevelError, want: "ERROR"},
		{name: "FATAL", level: LevelFatal, want: "FATAL"},
		{name: "other", level: 10, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.level.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLevelParseLevel(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want Level
	}{
		{name: "DEBUG", s: "DEBUG", want: LevelDebug},
		{name: "INFO", s: "INFO", want: LevelInfo},
		{name: "WARN", s: "WARN", want: LevelWarn},
		{name: "ERROR", s: "ERROR", want: LevelError},
		{name: "FATAL", s: "FATAL", want: LevelFatal},
		{name: "other", s: "other", want: LevelInfo},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseLevel(tt.s); got != tt.want {
				t.Errorf("ParseLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}
