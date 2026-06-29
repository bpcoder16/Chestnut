package aliyunoss

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	v2oss "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

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
		{
			name:        "pag",
			contentType: "application/x-pag",
			want:        ".pag",
		},
		{
			name:        "tencent pag with parameters",
			contentType: "application/vnd.tencent.pag; charset=binary",
			want:        ".pag",
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

func TestProcessObjectSaveAsPrivateBucketUsesSignedProcessURL(t *testing.T) {
	var processGETPath string
	var putPath string
	var putBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			processGETPath = r.URL.String()
			if got := r.URL.Query().Get("x-oss-process"); got != "image/watermark,text_dGVzdA" {
				t.Fatalf("x-oss-process = %q, want image/watermark,text_dGVzdA", got)
			}
			_, _ = w.Write([]byte("processed-image"))
		case http.MethodPut:
			putPath = r.URL.Path
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read put body: %v", err)
			}
			putBody = string(body)
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}))
	defer server.Close()

	endpoint := strings.TrimPrefix(server.URL, "http://")
	m := &Manager{
		scenes: map[string]*sceneEntry{
			"private-scene": {
				sceneConfig: &SceneConfig{
					BucketType:      OSSBucketTypePrivate,
					AccessKeyId:     "ak",
					AccessKeySecret: "sk",
					Endpoint:        endpoint,
					BucketName:      "private-bucket",
					Region:          "cn-hangzhou",
				},
				client: v2oss.NewClient(v2oss.LoadDefaultConfig().
					WithCredentialsProvider(credentials.NewStaticCredentialsProvider("ak", "sk")).
					WithRegion("cn-hangzhou").
					WithEndpoint(endpoint).
					WithDisableSSL(true).
					WithUsePathStyle(true)),
			},
		},
	}

	result, err := m.ProcessObjectSaveAs(context.Background(), "private-scene", "source.jpg", "target.jpg", "image/watermark,text_dGVzdA")
	if err != nil {
		t.Fatalf("ProcessObjectSaveAs() error = %v", err)
	}
	if !strings.HasPrefix(processGETPath, "/private-bucket/source.jpg?") {
		t.Fatalf("process GET path = %q, want /private-bucket/source.jpg?... ", processGETPath)
	}
	if putPath != "/private-bucket/target.jpg" {
		t.Fatalf("put path = %q, want /private-bucket/target.jpg", putPath)
	}
	if putBody != "processed-image" {
		t.Fatalf("put body = %q, want processed-image", putBody)
	}
	if result.Bucket != "private-bucket" || result.Object != "target.jpg" || result.FileSize != len("processed-image") || result.ProcessStatus != "OK" {
		t.Fatalf("result = %#v, want private-bucket target.jpg %d OK", result, len("processed-image"))
	}
}
