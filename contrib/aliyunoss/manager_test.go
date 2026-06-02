package aliyunoss

import "testing"

func TestExtFromContentTypeSupportsDocumentsAndVideo(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        string
	}{
		{
			name:        "pdf",
			contentType: "application/pdf",
			want:        ".pdf",
		},
		{
			name:        "mp4",
			contentType: "video/mp4",
			want:        ".mp4",
		},
		{
			name:        "mp4 with parameters",
			contentType: "video/mp4; charset=binary",
			want:        ".mp4",
		},
		{
			name:        "doc",
			contentType: "application/msword",
			want:        ".doc",
		},
		{
			name:        "docx",
			contentType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			want:        ".docx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtFromContentType(tt.contentType); got != tt.want {
				t.Fatalf("ExtFromContentType(%q) = %q, want %q", tt.contentType, got, tt.want)
			}
		})
	}
}
