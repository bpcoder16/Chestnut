package aliyunoss

import (
	"path"
	"strings"

	"github.com/bpcoder16/Chestnut/v4/appconfig/env"
	"github.com/bpcoder16/Chestnut/v4/core/utils"
)

const OSSBucketTypePublic = "public"
const OSSBucketTypePrivate = "private"
const DefaultOSSBucketType = OSSBucketTypePublic

type SceneConfig struct {
	BucketType      string   `yaml:"bucketType"`
	AccessKeyId     string   `yaml:"accessKeyId"`
	AccessKeySecret string   `yaml:"accessKeySecret"`
	Endpoint        string   `yaml:"endpoint"`
	StsEndpoint     string   `yaml:"stsEndpoint"`
	BucketName      string   `yaml:"bucketName"`
	CdnBaseURL      string   `yaml:"cdnBaseURL"`
	StsRoleArn      string   `yaml:"stsRoleArn"`
	StsSessionName  string   `yaml:"stsSessionName"`
	Region          string   `yaml:"region"`
	TargetDir       string   `yaml:"targetDir"`
	MaxBatchCount   int      `yaml:"maxBatchCount"`
	MaxSize         int64    `yaml:"maxSize"`
	AllowedExts     []string `yaml:"allowedExts"`
}

type OSSBucketConfig struct {
	AccessKeyId     string `yaml:"accessKeyId"`
	AccessKeySecret string `yaml:"accessKeySecret"`
	Endpoint        string `yaml:"endpoint"`
	StsEndpoint     string `yaml:"stsEndpoint"`
	BucketName      string `yaml:"bucketName"`
	CdnBaseURL      string `yaml:"cdnBaseURL"`
	StsRoleArn      string `yaml:"stsRoleArn"`
	StsSessionName  string `yaml:"stsSessionName"`
	Region          string `yaml:"region"`
}

type OSSConfig struct {
	Buckets map[string]*OSSBucketConfig `yaml:"buckets"`
	Scenes  map[string]*SceneConfig     `yaml:"scenes"`
}

func LoadOSSConfig(configPath string) *OSSConfig {
	var config OSSConfig
	if err := utils.ParseFile(configPath, &config); err != nil {
		panic("load OSS conf err: " + err.Error())
	}
	config.resolveSceneConfig()
	return &config
}

func (c *OSSConfig) resolveSceneConfig() {
	for _, scene := range c.Scenes {
		if scene.BucketType == "" {
			scene.BucketType = OSSBucketTypePublic
		}
		scene.TargetDir = prefixTargetDirWithRunMode(scene.TargetDir)
		if c.Buckets == nil {
			continue
		}
		bucket, ok := c.Buckets[scene.BucketType]
		if !ok || bucket == nil {
			continue
		}
		scene.fillEmptyBucketConfig(bucket)
	}
}

func prefixTargetDirWithRunMode(targetDir string) string {
	runMode := env.RunMode()
	targetDir = strings.Trim(targetDir, "/")
	if targetDir == "" {
		return runMode
	}
	// targetDir 的首段由 runMode 接管，避免环境配置遗漏导致不同环境写入同一 OSS 目录。
	dir := targetDir
	if first, rest, ok := strings.Cut(targetDir, "/"); ok && isRunModePrefix(first) {
		dir = rest
	} else if isRunModePrefix(targetDir) {
		dir = ""
	}
	return path.Join(runMode, dir)
}

func isRunModePrefix(prefix string) bool {
	switch prefix {
	case env.RunModeDebug, env.RunModeTest, env.RunModeRelease:
		return true
	default:
		return false
	}
}

func (c *SceneConfig) fillEmptyBucketConfig(bucket *OSSBucketConfig) {
	if c.AccessKeyId == "" {
		c.AccessKeyId = bucket.AccessKeyId
	}
	if c.AccessKeySecret == "" {
		c.AccessKeySecret = bucket.AccessKeySecret
	}
	if c.Endpoint == "" {
		c.Endpoint = bucket.Endpoint
	}
	if c.StsEndpoint == "" {
		c.StsEndpoint = bucket.StsEndpoint
	}
	if c.BucketName == "" {
		c.BucketName = bucket.BucketName
	}
	if c.CdnBaseURL == "" {
		c.CdnBaseURL = bucket.CdnBaseURL
	}
	if c.StsRoleArn == "" {
		c.StsRoleArn = bucket.StsRoleArn
	}
	if c.StsSessionName == "" {
		c.StsSessionName = bucket.StsSessionName
	}
	if c.Region == "" {
		c.Region = bucket.Region
	}
}
