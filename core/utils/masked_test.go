package utils

import "testing"

func TestMaskedName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "鬼", want: "鬼***"},
		{name: "鬼灭", want: "鬼***灭"},
		{name: "鬼灭之", want: "鬼***之"},
		{name: "鬼灭之刃", want: "鬼***刃"},
		{name: "B", want: "B***"},
		{name: "BP", want: "B***P"},
		{name: "BPC", want: "B***C"},
		{name: "BPCoder", want: "B***r"},
		{name: "Привет, мир", want: "П***р"},
		{name: "こんにちは、世界", want: "こ***界"},
		{name: "こんにちは、世界1", want: "こ***1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Masked(tt.name, "***"); got != tt.want {
				t.Errorf("MaskedName() = %v, want %v", got, tt.want)
			}
		})
	}
}
