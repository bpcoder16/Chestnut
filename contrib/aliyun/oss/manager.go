package oss

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/bpcoder16/Chestnut/v4/contrib/aliyun"
	"github.com/bpcoder16/Chestnut/v4/core/utils"
	"github.com/bpcoder16/Chestnut/v4/logit"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	sts "github.com/alibabacloud-go/sts-20150401/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	v2oss "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

const (
	BucketTypePublic  = "public"
	BucketTypePrivate = "private"
)

var imageHTTPClient = &http.Client{Timeout: 15 * time.Second}

var DefaultManager *Manager

type sceneEntry struct {
	sceneConfig  *aliyun.SceneConfig
	bucketConfig *aliyun.OSSBucketConfig
	client       *v2oss.Client
	stsClient    *sts.Client
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
	Region          string
	BucketName      string
}

func InitAliyunOSSManager(configPath string) {
	cfg := aliyun.LoadOSSConfig(configPath)
	m := &Manager{
		scenes: make(map[string]*sceneEntry, len(cfg.Scenes)),
	}
	for scene, sc := range cfg.Scenes {
		bucketType := sc.BucketType
		if bucketType == "" {
			bucketType = BucketTypePublic
		}
		sc.BucketType = bucketType
		bucketConfig := mergeBucketConfig(bucketType, sc, cfg.Buckets)
		ossCfg := v2oss.LoadDefaultConfig().
			WithCredentialsProvider(credentials.NewStaticCredentialsProvider(bucketConfig.AccessKeyId, bucketConfig.AccessKeySecret)).
			WithRegion(bucketConfig.Region)
		if bucketConfig.Endpoint != "" {
			ossCfg = ossCfg.WithEndpoint(strings.TrimPrefix(strings.TrimPrefix(bucketConfig.Endpoint, "https://"), "http://"))
		}
		ossClient := v2oss.NewClient(ossCfg)
		stsCfg := &openapi.Config{
			AccessKeyId:     tea.String(bucketConfig.AccessKeyId),
			AccessKeySecret: tea.String(bucketConfig.AccessKeySecret),
			RegionId:        tea.String(bucketConfig.Region),
		}
		if bucketConfig.StsEndpoint != "" {
			stsCfg.Endpoint = tea.String(strings.TrimPrefix(strings.TrimPrefix(bucketConfig.StsEndpoint, "https://"), "http://"))
		}
		stsClient, err := sts.NewClient(stsCfg)
		if err != nil {
			panic("oss: failed to create STS client for scene " + scene + ": " + err.Error())
		}
		m.scenes[scene] = &sceneEntry{
			sceneConfig:  sc,
			bucketConfig: bucketConfig,
			client:       ossClient,
			stsClient:    stsClient,
		}
	}
	DefaultManager = m
}

func mergeBucketConfig(bucketType string, sc *aliyun.SceneConfig, buckets map[string]*aliyun.OSSBucketConfig) *aliyun.OSSBucketConfig {
	if buckets != nil {
		if bc, ok := buckets[bucketType]; ok {
			return bc
		}
	}
	return &aliyun.OSSBucketConfig{
		AccessKeyId:     sc.AccessKeyId,
		AccessKeySecret: sc.AccessKeySecret,
		Endpoint:        sc.Endpoint,
		StsEndpoint:     sc.StsEndpoint,
		BucketName:      sc.BucketName,
		CdnBaseURL:      sc.CdnBaseURL,
		StsRoleArn:      sc.StsRoleArn,
		StsSessionName:  sc.StsSessionName,
		Region:          sc.Region,
	}
}

func (m *Manager) GetSceneConfig(scene string) (*aliyun.SceneConfig, error) {
	entry, ok := m.scenes[scene]
	if !ok {
		return nil, errors.New("oss: scene not found: " + scene)
	}
	return entry.sceneConfig, nil
}

func (m *Manager) GetSceneResolvedConfig(scene string) (*aliyun.SceneConfig, *aliyun.OSSBucketConfig, error) {
	entry, ok := m.scenes[scene]
	if !ok {
		return nil, nil, errors.New("oss: scene not found: " + scene)
	}
	return entry.sceneConfig, entry.bucketConfig, nil
}

const transferRetryCnt = 3

func extFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return filepath.Ext(u.Path)
}

func extFromContentType(contentType string) string {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return ""
	}
	switch mediaType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/bmp":
		return ".bmp"
	case "application/pdf":
		return ".pdf"
	default:
		return ""
	}
}

func buildTargetOSSPath(targetDir string, originURL string, contentType string) string {
	ext := extFromContentType(strings.ToLower(contentType))
	if ext == "" {
		ext = extFromURL(originURL)
	}
	return filepath.Join(targetDir, time.Now().Format("2006/01/02"), utils.UniqueID()+ext)
}

func (m *Manager) TransferImage(ctx context.Context, originURL, scene string) (ossPath string, err error) {
	entry, ok := m.scenes[scene]
	if !ok {
		return "", errors.New("oss: scene not found: " + scene)
	}
	httpResp, err := imageHTTPClient.Get(originURL)
	if err != nil {
		logit.Context(ctx).WarnW("oss.Manager.TransferImage", "http get failed", "err", err, "url", originURL)
		return "", err
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode >= 400 {
		err = errors.New("HTTPStatus:" + httpResp.Status)
		logit.Context(ctx).WarnW("oss.Manager.TransferImage", "http error", "err", err, "url", originURL)
		return "", err
	}
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		logit.Context(ctx).WarnW("oss.Manager.TransferImage", "read body failed", "err", err, "url", originURL)
		return "", err
	}
	ossPath = buildTargetOSSPath(entry.sceneConfig.TargetDir, originURL, httpResp.Header.Get("Content-Type"))
	for i := 0; i < transferRetryCnt; i++ {
		_, err = entry.client.PutObject(ctx, &v2oss.PutObjectRequest{
			Bucket: v2oss.Ptr(entry.bucketConfig.BucketName),
			Key:    v2oss.Ptr(ossPath),
			Body:   bytes.NewReader(body),
			//ContentType: v2oss.Ptr(httpResp.Header.Get("Content-Type")),
		})
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
	return m.genStsToken(scene, durationSeconds, "")
}

func (m *Manager) GenStsTokenForObjects(scene string, objectKeys []string, durationSeconds int64) (*StsCredentials, error) {
	entry, ok := m.scenes[scene]
	if !ok {
		return nil, errors.New("oss: scene not found: " + scene)
	}
	policy, err := buildPutObjectPolicy(entry.bucketConfig.BucketName, objectKeys)
	if err != nil {
		return nil, err
	}
	return m.genStsToken(scene, durationSeconds, policy)
}

func (m *Manager) genStsToken(scene string, durationSeconds int64, policy string) (*StsCredentials, error) {
	entry, ok := m.scenes[scene]
	if !ok {
		return nil, errors.New("oss: scene not found: " + scene)
	}
	req := &sts.AssumeRoleRequest{
		// RAM角色的RamRoleArn。
		RoleArn: tea.String(entry.bucketConfig.StsRoleArn),
		// 指定自定义角色会话名称
		RoleSessionName: tea.String(entry.bucketConfig.StsSessionName),
		// 指定STS临时访问凭证过期时间
		DurationSeconds: tea.Int64(durationSeconds),
	}
	if policy != "" {
		req.Policy = tea.String(policy)
	}
	resp, err := entry.stsClient.AssumeRole(req)
	if err != nil {
		return nil, err
	}
	cred := resp.Body.Credentials
	return &StsCredentials{
		AccessKeyId:     tea.StringValue(cred.AccessKeyId),
		AccessKeySecret: tea.StringValue(cred.AccessKeySecret),
		SecurityToken:   tea.StringValue(cred.SecurityToken),
		Expiration:      tea.StringValue(cred.Expiration),
		Endpoint:        entry.bucketConfig.Endpoint,
		Region:          entry.bucketConfig.Region,
		BucketName:      entry.bucketConfig.BucketName,
	}, nil
}

type putObjectPolicy struct {
	Version   string               `json:"Version"`
	Statement []putObjectStatement `json:"Statement"`
}

type putObjectStatement struct {
	Effect   string   `json:"Effect"`
	Action   []string `json:"Action"`
	Resource []string `json:"Resource"`
}

func buildPutObjectPolicy(bucketName string, objectKeys []string) (string, error) {
	resources := make([]string, 0, len(objectKeys))
	for _, objectKey := range objectKeys {
		if objectKey == "" {
			return "", errors.New("oss: empty object key")
		}
		resources = append(resources, "acs:oss:*:*:"+bucketName+"/"+objectKey)
	}
	body, err := json.Marshal(putObjectPolicy{
		Version: "1",
		Statement: []putObjectStatement{
			{
				Effect:   "Allow",
				Action:   []string{"oss:PutObject"},
				Resource: resources,
			},
		},
	})
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (m *Manager) IsObjectExist(ctx context.Context, scene, objectKey string) (bool, error) {
	entry, ok := m.scenes[scene]
	if !ok {
		return false, errors.New("oss: scene not found: " + scene)
	}
	return entry.client.IsObjectExist(ctx, entry.bucketConfig.BucketName, objectKey)
}

func (m *Manager) SignGetObjectURL(ctx context.Context, scene, objectKey string, expiredInSec int64) (*v2oss.PresignResult, error) {
	entry, ok := m.scenes[scene]
	if !ok {
		return nil, errors.New("oss: scene not found: " + scene)
	}
	return entry.client.Presign(ctx, &v2oss.GetObjectRequest{
		Bucket: v2oss.Ptr(entry.bucketConfig.BucketName),
		Key:    v2oss.Ptr(objectKey),
	}, v2oss.PresignExpires(time.Duration(expiredInSec)*time.Second))
}
