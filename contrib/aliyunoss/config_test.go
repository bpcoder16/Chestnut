package aliyunoss

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOSSConfigResolveSceneConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "oss.yaml")
	configBody := []byte(`
buckets:
  public:
    accessKeyId: bucket-ak
    accessKeySecret: bucket-sk
    endpoint: bucket-endpoint
    stsEndpoint: bucket-sts-endpoint
    bucketName: bucket-name
    cdnBaseURL: bucket-cdn
    stsRoleArn: bucket-role
    stsSessionName: bucket-session
    region: bucket-region
scenes:
  upload:
    targetDir: uploads
    accessKeyId: scene-ak
    bucketName: scene-bucket
`)
	if err := os.WriteFile(configPath, configBody, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg := LoadOSSConfig(configPath)
	scene := cfg.Scenes["upload"]
	if scene == nil {
		t.Fatal("scene upload not loaded")
	}
	if scene.BucketType != DefaultOSSBucketType {
		t.Fatalf("BucketType = %q, want %q", scene.BucketType, DefaultOSSBucketType)
	}
	if scene.AccessKeyId != "scene-ak" {
		t.Fatalf("AccessKeyId = %q, want scene-ak", scene.AccessKeyId)
	}
	if scene.BucketName != "scene-bucket" {
		t.Fatalf("BucketName = %q, want scene-bucket", scene.BucketName)
	}
	if scene.AccessKeySecret != "bucket-sk" {
		t.Fatalf("AccessKeySecret = %q, want bucket-sk", scene.AccessKeySecret)
	}
	if scene.Endpoint != "bucket-endpoint" {
		t.Fatalf("Endpoint = %q, want bucket-endpoint", scene.Endpoint)
	}
	if scene.StsEndpoint != "bucket-sts-endpoint" {
		t.Fatalf("StsEndpoint = %q, want bucket-sts-endpoint", scene.StsEndpoint)
	}
	if scene.CdnBaseURL != "bucket-cdn" {
		t.Fatalf("CdnBaseURL = %q, want bucket-cdn", scene.CdnBaseURL)
	}
	if scene.StsRoleArn != "bucket-role" {
		t.Fatalf("StsRoleArn = %q, want bucket-role", scene.StsRoleArn)
	}
	if scene.StsSessionName != "bucket-session" {
		t.Fatalf("StsSessionName = %q, want bucket-session", scene.StsSessionName)
	}
	if scene.Region != "bucket-region" {
		t.Fatalf("Region = %q, want bucket-region", scene.Region)
	}
}
