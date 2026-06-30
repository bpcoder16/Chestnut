package aliyunoss

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bpcoder16/Chestnut/v4/core/utils"
	"github.com/bpcoder16/Chestnut/v4/logit"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	sts "github.com/alibabacloud-go/sts-20150401/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	v2oss "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

const (
	ObjectFormatSourceContentType = "content_type"
	ObjectFormatSourceObjectKey   = "object_key"
)

var imageHTTPClient = &http.Client{Timeout: 15 * time.Second}

var DefaultManager *Manager

type sceneEntry struct {
	sceneConfig *SceneConfig
	client      *v2oss.Client
	stsClient   *sts.Client
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

// ObjectFormat OSS 文件格式信息。
type ObjectFormat struct {
	// Format 文件格式，不带点，例如 jpg、jpeg、png、pdf；优先根据 OSS Content-Type 判断，无法判断时使用 object key 后缀。
	Format string
	// ContentType OSS 对象的 Content-Type，例如 image/jpeg、application/pdf。
	ContentType string
	// Source 格式识别来源：content_type 表示来自 Content-Type，object_key 表示来自 object key 后缀。
	Source string
}

// TextWatermarkOptions 文字水印处理配置。
type TextWatermarkOptions struct {
	// TargetObjectKey 处理后图片保存到 OSS 的 object key，例如 processed/2026/05/09/demo.jpg。
	TargetObjectKey string

	// Transparency 文字水印透明度，对应 OSS 参数 t，取值范围遵循 OSS 图片处理规则，示例：90。
	Transparency int
	// Position 水印位置，对应 OSS 参数 g，示例：se 表示右下。
	Position string
	// X 水平边距，对应 OSS 参数 x，单位 px，示例：10。
	X int
	// Y 垂直边距，对应 OSS 参数 y，单位 px，示例：10。
	Y int
	// 指定是否将图片水印或文字水印铺满原图 1：将图片水印或文字水印铺满原图 0（默认值）：不将图片水印或文字水印铺满全图
	Fill int

	// Text 水印文字内容，方法内部会做 URL-safe Base64 编码；OSS 限制最大 64 个字符，中文约 20 个字。
	Text string
	// Font 字体名称，方法内部会做 URL-safe Base64 编码；空值使用 OSS 默认字体 wqy-zenhei，示例：wqy-zenhei。
	Font string
	// Color 文字颜色，格式为 RRGGBB 或 #RRGGBB，方法内部会移除 # 后传给 OSS，示例：FFFFFF。
	Color string
	// Size 字体大小，单位 px；空值使用 OSS 默认值，示例：40。
	Size int
	// Shadow 文字阴影透明度，对应 OSS 参数 shadow，取值范围 0-100，示例：50。
	Shadow int
	// Rotate 指定文字顺时针旋转角度 取值范围 0-360，示例：50
	Rotate int
}

func InitAliyunOSSManager(configPath string) {
	cfg := LoadOSSConfig(configPath)
	m := &Manager{
		scenes: make(map[string]*sceneEntry, len(cfg.Scenes)),
	}
	for scene, sc := range cfg.Scenes {
		ossCfg := v2oss.LoadDefaultConfig().
			WithCredentialsProvider(credentials.NewStaticCredentialsProvider(sc.AccessKeyId, sc.AccessKeySecret)).
			WithRegion(sc.Region)
		if sc.Endpoint != "" {
			ossCfg = ossCfg.WithEndpoint(strings.TrimPrefix(strings.TrimPrefix(sc.Endpoint, "https://"), "http://"))
		}
		ossClient := v2oss.NewClient(ossCfg)
		stsCfg := &openapi.Config{
			AccessKeyId:     tea.String(sc.AccessKeyId),
			AccessKeySecret: tea.String(sc.AccessKeySecret),
			RegionId:        tea.String(sc.Region),
		}
		if sc.StsEndpoint != "" {
			stsCfg.Endpoint = tea.String(strings.TrimPrefix(strings.TrimPrefix(sc.StsEndpoint, "https://"), "http://"))
		}
		stsClient, err := sts.NewClient(stsCfg)
		if err != nil {
			panic("oss: failed to create STS client for scene " + scene + ": " + err.Error())
		}
		m.scenes[scene] = &sceneEntry{
			sceneConfig: sc,
			client:      ossClient,
			stsClient:   stsClient,
		}
	}
	DefaultManager = m
}

func (m *Manager) GetSceneConfig(scene string) (*SceneConfig, error) {
	entry, ok := m.scenes[scene]
	if !ok {
		return nil, errors.New("oss: scene not found: " + scene)
	}
	return entry.sceneConfig, nil
}

const transferRetryCnt = 3
const processSignedURLExpireSeconds int64 = 10 * 60

func ExtFromContentType(contentType string) string {
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
	case "video/mp4":
		return ".mp4"
	case "application/msword":
		return ".doc"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return ".docx"
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return ".xlsx"
	case "application/vnd.ms-excel":
		return ".xls"
	case "text/csv", "application/csv":
		return ".csv"
	case "application/json", "text/json":
		return ".json"
	case "application/x-pag", "application/vnd.tencent.pag":
		return ".pag"
	case "application/vnd.android.package-archive":
		return ".apk"
	case "text/x-patch", "text/x-diff", "application/x-patch", "application/x-diff":
		return ".patch"
	default:
		return ""
	}
}

func formatFromExt(ext string) string {
	return strings.ToLower(strings.TrimPrefix(ext, "."))
}

func normalizeObjectExt(fileExt string) string {
	fileExt = strings.TrimSpace(fileExt)
	if fileExt == "" {
		return ""
	}
	fileExt = "." + strings.TrimPrefix(strings.ToLower(fileExt), ".")
	if fileExt == "." || strings.ContainsAny(fileExt, `/\`) {
		return ""
	}
	return fileExt
}

func transferObjectExt(contentType, fileExt string) string {
	if ext := normalizeObjectExt(fileExt); ext != "" {
		return ext
	}
	return ExtFromContentType(strings.ToLower(contentType))
}

// BuildTargetOSSPath 根据场景配置和 Content-Type 生成 OSS object key。
func BuildTargetOSSPath(scene string, contentType string, extraDirs ...string) (string, error) {
	return DefaultManager.BuildTargetOSSPath(scene, contentType, extraDirs...)
}

// BuildTargetOSSPath 根据场景配置和 Content-Type 生成 OSS object key。
func (m *Manager) BuildTargetOSSPath(scene string, contentType string, extraDirs ...string) (string, error) {
	return m.buildTargetOSSPathWithObjectExt(scene, ExtFromContentType(strings.ToLower(contentType)), extraDirs...)
}

// BuildTargetOSSPathWithExt 根据场景配置和文件后缀生成 OSS object key。
func BuildTargetOSSPathWithExt(scene string, fileExt string, extraDirs ...string) (string, error) {
	return DefaultManager.BuildTargetOSSPathWithExt(scene, fileExt, extraDirs...)
}

// BuildTargetOSSPathWithExt 根据场景配置和文件后缀生成 OSS object key。
func (m *Manager) BuildTargetOSSPathWithExt(scene string, fileExt string, extraDirs ...string) (string, error) {
	return m.buildTargetOSSPathWithObjectExt(scene, normalizeObjectExt(fileExt), extraDirs...)
}

func (m *Manager) buildTargetOSSPathWithObjectExt(scene string, ext string, extraDirs ...string) (string, error) {
	entry, ok := m.scenes[scene]
	if !ok {
		return "", errors.New("oss: scene not found: " + scene)
	}
	targetDir := entry.sceneConfig.TargetDir
	for _, extraDir := range extraDirs {
		extraDir = strings.Trim(extraDir, "/")
		if extraDir != "" && extraDir != "." {
			targetDir = filepath.Join(targetDir, extraDir)
		}
	}
	return buildTargetOSSPathWithExt(targetDir, ext), nil
}

func buildTargetOSSPathWithExt(targetDir string, ext string) string {
	return filepath.Join(targetDir, time.Now().Format("2006/01/02"), utils.UniqueID()+ext)
}

func (m *Manager) TransferImage(ctx context.Context, originURL, scene string, extraDirs ...string) (targetObjectKey string, contentType string, fileSize int64, err error) {
	return m.transferOSSObject(ctx, originURL, scene, "", "oss.Manager.TransferImage", extraDirs...)
}

// TransferOSSObject 下载远端文件并上传到 OSS，fileExt 非空时优先作为目标 object key 后缀。
func (m *Manager) TransferOSSObject(ctx context.Context, originURL, scene, fileExt string, extraDirs ...string) (targetObjectKey string, contentType string, fileSize int64, err error) {
	return m.transferOSSObject(ctx, originURL, scene, fileExt, "oss.Manager.TransferOSSObject", extraDirs...)
}

func (m *Manager) transferOSSObject(ctx context.Context, originURL, scene, fileExt, logField string, extraDirs ...string) (targetObjectKey string, contentType string, fileSize int64, err error) {
	entry, ok := m.scenes[scene]
	if !ok {
		err = errors.New("oss: scene not found: " + scene)
		return
	}
	httpResp, err := imageHTTPClient.Get(originURL)
	if err != nil {
		logit.Context(ctx).WarnW(logField, "http get failed", "err", err, "url", originURL)
		return
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode >= 400 {
		err = errors.New("HTTPStatus:" + httpResp.Status)
		logit.Context(ctx).WarnW(logField, "http error", "err", err, "url", originURL)
		return
	}
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		logit.Context(ctx).WarnW(logField, "read body failed", "err", err, "url", originURL)
		return
	}

	contentType = httpResp.Header.Get("Content-Type")
	fileSize = int64(len(body))

	targetObjectKey, err = m.buildTargetOSSPathWithObjectExt(scene, transferObjectExt(contentType, fileExt), extraDirs...)
	if err != nil {
		logit.Context(ctx).WarnW(logField, "BuildTargetOSSPath failed", "err", err)
		return
	}

	for i := 0; i < transferRetryCnt; i++ {
		_, err = entry.client.PutObject(ctx, &v2oss.PutObjectRequest{
			Bucket: v2oss.Ptr(entry.sceneConfig.BucketName),
			Key:    v2oss.Ptr(targetObjectKey),
			Body:   bytes.NewReader(body),
		})
		if err == nil {
			break
		}
	}
	if err != nil {
		logit.Context(ctx).WarnW(logField, "upload failed", "err", err)
	}
	return
}

func (m *Manager) ProcessObjectSaveAs(ctx context.Context, scene, sourceObjectKey, targetObjectKey, process string) (*v2oss.ProcessObjectResult, error) {
	entry, ok := m.scenes[scene]
	if !ok {
		return nil, errors.New("oss: scene not found: " + scene)
	}
	if sourceObjectKey == "" {
		return nil, errors.New("oss: empty source object key")
	}
	if targetObjectKey == "" {
		return nil, errors.New("oss: empty target object key")
	}
	if process == "" {
		return nil, errors.New("oss: empty process")
	}
	process = strings.TrimSuffix(process, "|")
	if entry.sceneConfig.BucketType == OSSBucketTypePrivate {
		// 私有 bucket 不能直接用未签名的图片处理 URL，先生成带 x-oss-process 的签名 URL 获取处理结果，再保存到目标对象。
		return m.processObjectSaveAsPrivate(ctx, entry, sourceObjectKey, targetObjectKey, process)
	}
	process = process + "|sys/saveas,o_" + base64.URLEncoding.EncodeToString([]byte(targetObjectKey))
	return entry.client.ProcessObject(ctx, &v2oss.ProcessObjectRequest{
		Bucket:  v2oss.Ptr(entry.sceneConfig.BucketName),
		Key:     v2oss.Ptr(sourceObjectKey),
		Process: v2oss.Ptr(process),
	})
}

func (m *Manager) processObjectSaveAsPrivate(ctx context.Context, entry *sceneEntry, sourceObjectKey, targetObjectKey, process string) (*v2oss.ProcessObjectResult, error) {
	presignResult, err := entry.client.Presign(ctx, &v2oss.GetObjectRequest{
		Bucket:  v2oss.Ptr(entry.sceneConfig.BucketName),
		Key:     v2oss.Ptr(sourceObjectKey),
		Process: v2oss.Ptr(process),
	}, v2oss.PresignExpires(time.Duration(processSignedURLExpireSeconds)*time.Second))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, presignResult.URL, nil)
	if err != nil {
		return nil, err
	}
	for key, value := range presignResult.SignedHeaders {
		req.Header.Set(key, value)
	}
	httpResp, err := imageHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode >= http.StatusBadRequest {
		return nil, errors.New("HTTPStatus:" + httpResp.Status)
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}
	putReq := &v2oss.PutObjectRequest{
		Bucket: v2oss.Ptr(entry.sceneConfig.BucketName),
		Key:    v2oss.Ptr(targetObjectKey),
		Body:   bytes.NewReader(body),
	}
	if contentType := httpResp.Header.Get("Content-Type"); contentType != "" {
		putReq.ContentType = v2oss.Ptr(contentType)
	}
	if _, err = entry.client.PutObject(ctx, putReq); err != nil {
		return nil, err
	}

	return &v2oss.ProcessObjectResult{
		Bucket:        entry.sceneConfig.BucketName,
		FileSize:      len(body),
		Object:        targetObjectKey,
		ProcessStatus: "OK",
		ResultCommon: v2oss.ResultCommon{
			Status:     httpResp.Status,
			StatusCode: httpResp.StatusCode,
			Headers:    httpResp.Header,
		},
	}, nil
}

func (m *Manager) ProcessTextWatermarkSaveAs(ctx context.Context, scene, sourceObjectKey string, opts TextWatermarkOptions) (*v2oss.ProcessObjectResult, error) {
	if opts.Text == "" {
		return nil, errors.New("oss: empty watermark text")
	}
	process := "image/watermark,text_" + base64.RawURLEncoding.EncodeToString([]byte(opts.Text))
	if opts.Transparency > 0 {
		process += ",t_" + strconv.Itoa(opts.Transparency)
	}
	if opts.Position != "" {
		process += ",g_" + opts.Position
	}
	if opts.X > 0 {
		process += ",x_" + strconv.Itoa(opts.X)
	}
	if opts.Y > 0 {
		process += ",y_" + strconv.Itoa(opts.Y)
	}
	if opts.Fill == 1 {
		process += ",fill_" + strconv.Itoa(opts.Fill)
	}

	if opts.Font != "" {
		process += ",type_" + base64.RawURLEncoding.EncodeToString([]byte(opts.Font))
	}
	if opts.Color != "" {
		process += ",color_" + strings.TrimPrefix(opts.Color, "#")
	}
	if opts.Size > 0 {
		process += ",size_" + strconv.Itoa(opts.Size)
	}
	if opts.Shadow > 0 {
		process += ",shadow_" + strconv.Itoa(opts.Shadow)
	}
	if opts.Rotate > 0 {
		process += ",rotate_" + strconv.Itoa(opts.Rotate)
	}

	return m.ProcessObjectSaveAs(ctx, scene, sourceObjectKey, opts.TargetObjectKey, process)
}

func (m *Manager) GetObjectFormat(ctx context.Context, scene, objectKey string) (*ObjectFormat, error) {
	entry, ok := m.scenes[scene]
	if !ok {
		return nil, errors.New("oss: scene not found: " + scene)
	}
	if objectKey == "" {
		return nil, errors.New("oss: empty object key")
	}
	result, err := entry.client.HeadObject(ctx, &v2oss.HeadObjectRequest{
		Bucket: v2oss.Ptr(entry.sceneConfig.BucketName),
		Key:    v2oss.Ptr(objectKey),
	})
	if err != nil {
		return nil, err
	}
	contentType := ""
	if result.ContentType != nil {
		contentType = *result.ContentType
	}
	if ext := ExtFromContentType(strings.ToLower(contentType)); ext != "" {
		return &ObjectFormat{
			Format:      strings.TrimPrefix(ext, "."),
			ContentType: contentType,
			Source:      ObjectFormatSourceContentType,
		}, nil
	}
	return &ObjectFormat{
		Format:      formatFromExt(filepath.Ext(objectKey)),
		ContentType: contentType,
		Source:      ObjectFormatSourceObjectKey,
	}, nil
}

func (m *Manager) GenStsToken(scene string, durationSeconds int64) (*StsCredentials, error) {
	return m.genStsToken(scene, durationSeconds, "")
}

func (m *Manager) GenStsTokenForObjects(scene string, objectKeys []string, durationSeconds int64) (*StsCredentials, error) {
	entry, ok := m.scenes[scene]
	if !ok {
		return nil, errors.New("oss: scene not found: " + scene)
	}
	policy, err := buildPutObjectPolicy(entry.sceneConfig.BucketName, objectKeys)
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
		RoleArn: tea.String(entry.sceneConfig.StsRoleArn),
		// 指定自定义角色会话名称
		RoleSessionName: tea.String(entry.sceneConfig.StsSessionName),
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
		Endpoint:        entry.sceneConfig.Endpoint,
		Region:          entry.sceneConfig.Region,
		BucketName:      entry.sceneConfig.BucketName,
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
	return entry.client.IsObjectExist(ctx, entry.sceneConfig.BucketName, objectKey)
}

func (m *Manager) DeleteObject(ctx context.Context, scene, objectKey string) error {
	entry, ok := m.scenes[scene]
	if !ok {
		return errors.New("oss: scene not found: " + scene)
	}
	_, err := entry.client.DeleteObject(ctx, &v2oss.DeleteObjectRequest{
		Bucket: v2oss.Ptr(entry.sceneConfig.BucketName),
		Key:    v2oss.Ptr(objectKey),
	})
	return err
}

func (m *Manager) SignGetObjectURL(ctx context.Context, scene, objectKey string, expiredInSec int64) (*v2oss.PresignResult, error) {
	entry, ok := m.scenes[scene]
	if !ok {
		return nil, errors.New("oss: scene not found: " + scene)
	}
	return entry.client.Presign(ctx, &v2oss.GetObjectRequest{
		Bucket: v2oss.Ptr(entry.sceneConfig.BucketName),
		Key:    v2oss.Ptr(objectKey),
	}, v2oss.PresignExpires(time.Duration(expiredInSec)*time.Second))
}
