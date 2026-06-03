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
		{
			name:        "json",
			contentType: "application/json",
			want:        ".json",
		},
		{
			name:        "text json with parameters",
			contentType: "text/json; charset=utf-8",
			want:        ".json",
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

func TestTransferObjectExtPrefersManualExt(t *testing.T) {
	if got := transferObjectExt("image/png", "jpg"); got != ".jpg" {
		t.Fatalf("transferObjectExt() = %q, want %q", got, ".jpg")
	}
}

func TestTransferObjectExtFallsBackToContentType(t *testing.T) {
	if got := transferObjectExt("image/png", ""); got != ".png" {
		t.Fatalf("transferObjectExt() = %q, want %q", got, ".png")
	}
}
