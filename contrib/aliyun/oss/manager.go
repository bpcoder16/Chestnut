package oss

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"

	"github.com/bpcoder16/Chestnut/v4/contrib/aliyun"
	"github.com/bpcoder16/Chestnut/v4/core/utils"
	"github.com/bpcoder16/Chestnut/v4/default/resty"
	"github.com/bpcoder16/Chestnut/v4/logit"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	sts "github.com/alibabacloud-go/sts-20150401/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	gosdk "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var DefaultManager *Manager

type sceneEntry struct {
	config    *aliyun.SceneConfig
	bucket    *gosdk.Bucket
	stsClient *sts.Client
}

type Manager struct {
	scenes map[string]*sceneEntry
}

type StsCredentials struct {
	AccessKeyId     string
	AccessKeySecret string
	SecurityToken   string
	Expiration      string
	Endpoint        string
	BucketName      string
	TargetDir       string
}

func InitAliyunOSSManager(configPath string) {
	cfg := aliyun.LoadOSSConfig(configPath)
	m := &Manager{
		scenes: make(map[string]*sceneEntry, len(cfg.Scenes)),
	}
	for scene, sc := range cfg.Scenes {
		ossClient, err := gosdk.New(sc.Endpoint, sc.AccessKeyId, sc.AccessKeySecret)
		if err != nil {
			panic("oss: failed to create client for scene " + scene + ": " + err.Error())
		}
		bucket, err := ossClient.Bucket(sc.BucketName)
		if err != nil {
			panic("oss: failed to get bucket for scene " + scene + ": " + err.Error())
		}
		stsClient, err := sts.NewClient(&openapi.Config{
			AccessKeyId:     dara.String(sc.AccessKeyId),
			AccessKeySecret: dara.String(sc.AccessKeySecret),
			RegionId:        dara.String(sc.Region),
		})
		if err != nil {
			panic("oss: failed to create STS client for scene " + scene + ": " + err.Error())
		}
		m.scenes[scene] = &sceneEntry{
			config:    sc,
			bucket:    bucket,
			stsClient: stsClient,
		}
	}
	DefaultManager = m
}

func (m *Manager) GetBucket(scene string) (*gosdk.Bucket, error) {
	entry, ok := m.scenes[scene]
	if !ok {
		return nil, errors.New("oss: scene not found: " + scene)
	}
	return entry.bucket, nil
}

func (m *Manager) GetSceneConfig(scene string) (*aliyun.SceneConfig, error) {
	entry, ok := m.scenes[scene]
	if !ok {
		return nil, errors.New("oss: scene not found: " + scene)
	}
	return entry.config, nil
}

const transferRetryCnt = 3

func buildTargetOSSPath(targetDir, originURL string) string {
	ext := filepath.Ext(originURL)
	return filepath.Join(targetDir, utils.UniqueID()+ext)
}

func (m *Manager) TransferImage(ctx context.Context, originURL, scene string) (ossPath string, err error) {
	entry, ok := m.scenes[scene]
	if !ok {
		return "", errors.New("oss: scene not found: " + scene)
	}
	resp, err := resty.Client().R().Get(originURL)
	if err != nil {
		logit.Context(ctx).WarnW("oss.Manager.TransferImage", "http get failed", "err", err, "url", originURL)
		return "", err
	}
	if resp.IsError() {
		err = errors.New("HTTPStatus:" + resp.Status())
		logit.Context(ctx).WarnW("oss.Manager.TransferImage", "http error", "err", err, "url", originURL)
		return "", err
	}
	ossPath = buildTargetOSSPath(entry.config.TargetDir, originURL)
	imageData := bytes.NewReader(resp.Body())
	for i := 0; i < transferRetryCnt; i++ {
		_, _ = imageData.Seek(0, 0)
		err = entry.bucket.PutObject(ossPath, imageData)
		if err == nil {
			break
		}
	}
	if err != nil {
		logit.Context(ctx).WarnW("oss.Manager.TransferImage", "upload failed", "err", err, "ossPath", ossPath)
		ossPath = ""
	}
	return
}

func (m *Manager) GenStsToken(scene string, durationSeconds int64) (*StsCredentials, error) {
	entry, ok := m.scenes[scene]
	if !ok {
		return nil, errors.New("oss: scene not found: " + scene)
	}
	req := &sts.AssumeRoleRequest{
		RoleArn:         dara.String(entry.config.StsRoleArn),
		RoleSessionName: dara.String(entry.config.StsSessionName),
		DurationSeconds: dara.Int64(durationSeconds),
	}
	resp, err := entry.stsClient.AssumeRole(req)
	if err != nil {
		return nil, err
	}
	cred := resp.Body.Credentials
	return &StsCredentials{
		AccessKeyId:     dara.StringValue(cred.AccessKeyId),
		AccessKeySecret: dara.StringValue(cred.AccessKeySecret),
		SecurityToken:   dara.StringValue(cred.SecurityToken),
		Expiration:      dara.StringValue(cred.Expiration),
		Endpoint:        entry.config.Endpoint,
		BucketName:      entry.config.BucketName,
		TargetDir:       entry.config.TargetDir,
	}, nil
}
