package zaplogger

import (
	"testing"
	"time"
)

const testRetentionDays = 7

func TestLogRetentionMaxAge(t *testing.T) {
	tests := []struct {
		name          string
		retentionDays int
		want          time.Duration
	}{
		{
			name: "not configured",
			want: defaultLogRetentionDays * 24 * time.Hour,
		},
		{
			name:          "configured days",
			retentionDays: testRetentionDays,
			want:          testRetentionDays * 24 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := logRetentionMaxAge(tt.retentionDays); got != tt.want {
				t.Fatalf("logRetentionMaxAge() = %s, want %s", got, tt.want)
			}
		})
	}
}
