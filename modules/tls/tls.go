package tls

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/bpcoder16/Chestnut/v2/core/utils"
)

type KeyType string

const (
	KeyTypeRSA     KeyType = "rsa"
	KeyTypeECDSA   KeyType = "ecdsa"
	KeyTypeEd25519 KeyType = "ed25519"
)

type ECDSACurve string

const (
	CurveP256 ECDSACurve = "P256"
	CurveP384 ECDSACurve = "P384"
	CurveP521 ECDSACurve = "P521"
)

type PrivateKeyConfig struct {
	KeyType KeyType
	KeySize int
	Curve   ECDSACurve
}

func ensureDir(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败 %s: %w", dir, err)
	}
	return nil
}

func generatePrivateKey(config *PrivateKeyConfig) (any, error) {
	switch config.KeyType {
	case KeyTypeRSA:
		return rsa.GenerateKey(rand.Reader, config.KeySize)
	case KeyTypeECDSA:
		var ellipticCurve elliptic.Curve
		switch config.Curve {
		case CurveP256:
			ellipticCurve = elliptic.P256()
		case CurveP384:
			ellipticCurve = elliptic.P384()
		case CurveP521:
			ellipticCurve = elliptic.P521()
		default:
			ellipticCurve = elliptic.P256() // 默认使用 P-256
		}
		return ecdsa.GenerateKey(ellipticCurve, rand.Reader)
	case KeyTypeEd25519:
		_, privateKey, err := ed25519.GenerateKey(rand.Reader)
		return privateKey, err
	default:
		return nil, fmt.Errorf("不支持的密钥类型: %s", config.KeyType)
	}
}

func getPublicKey(privateKey any) (any, error) {
	switch key := privateKey.(type) {
	case *rsa.PrivateKey:
		return &key.PublicKey, nil
	case *ecdsa.PrivateKey:
		return &key.PublicKey, nil
	case ed25519.PrivateKey:
		return key.Public(), nil
	default:
		return nil, fmt.Errorf("不支持的CA私钥类型: %T", privateKey)
	}
}

func generateSerialNumber() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	return rand.Int(rand.Reader, serialNumberLimit)
}

type CAConfig struct {
	Dir              string
	Filename         string
	CommonName       string   // 证书持有者的通用名称
	Organization     []string // 证书信任链中显示的组织信息
	Country          []string // 国家代码 (ISO 3166-1 alpha-2)
	Province         []string // 省份或州
	Locality         []string // 城市或地区
	ValidDays        int64
	PrivateKeyConfig *PrivateKeyConfig
}

func NewDefaultCAConfig(dir, filename, commonName string, validDays int64) *CAConfig {
	return &CAConfig{
		Dir:          dir,
		Filename:     filename,
		CommonName:   commonName,
		Organization: []string{"YM&MH Studio"},
		Country:      []string{"CN"},
		Province:     []string{"Liaoning"},
		Locality:     []string{"Shenyang"},
		ValidDays:    validDays,
		PrivateKeyConfig: &PrivateKeyConfig{
			KeyType: KeyTypeECDSA,
			Curve:   CurveP256,
		},
	}
}

// GenerateCA 创建 CA 证书
//
//	config := CAConfig{
//	    Dir:          "/path/to/ca",
//	    FileName:   "goumang-master",
//	    CommonName:   "GouMang Master Root CA",
//	    Organization: []string{"YM&MH Studio"},
//	    Country:      []string{"CN"},
//	    Province:     []string{"Liaoning"},
//	    Locality:     []string{"Shenyang"},
//		ValidDays:    validDays,
//	    PrivateKeyConfig: PrivateKeyConfig{
//	        KeyType: KeyTypeECDSA,
//	        Curve:   CurveP256,
//	    },
//	}
func GenerateCA(config *CAConfig) (caCert *x509.Certificate, caPrivateKey any, err error) {
	if err = ensureDir(config.Dir); err != nil {
		return
	}

	// 生成 CA 的公钥和私钥
	caPrivateKey, err = generatePrivateKey(config.PrivateKeyConfig)
	if err != nil {
		err = fmt.Errorf("生成 CA 私钥失败: %w", err)
		return
	}
	caPublicKey, err := getPublicKey(caPrivateKey)
	if err != nil {
		return
	}

	// 生成 CA 证书
	serialNumber, err := generateSerialNumber()
	if err != nil {
		err = fmt.Errorf("生成序列号失败: %w", err)
		return
	}
	pkixName := pkix.Name{
		CommonName:   config.CommonName,
		Organization: config.Organization,
		Country:      config.Country,
		Province:     config.Province,
		Locality:     config.Locality,
	}
	caCert = &x509.Certificate{
		SerialNumber:          serialNumber,
		Issuer:                pkixName,
		Subject:               pkixName,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Duration(config.ValidDays) * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	// 自签名的 CA 根证书
	caCertDER, err := x509.CreateCertificate(rand.Reader, caCert, caCert, caPublicKey, caPrivateKey)
	if err != nil {
		err = fmt.Errorf("创建 CA 证书数据失败: %w", err)
		return
	}
	caCertPath := filepath.Join(config.Dir, config.Filename+".crt")
	caCertFile, err := os.Create(caCertPath)
	if err != nil {
		err = fmt.Errorf("创建 CA 证书文件失败: %w", err)
		return
	}
	defer func() {
		_ = caCertFile.Close()
	}()
	if err = pem.Encode(caCertFile, &pem.Block{Type: "CERTIFICATE", Bytes: caCertDER}); err != nil {
		err = fmt.Errorf("写入 CA 证书失败: %w", err)
		return
	}

	// 生成 CA 私钥文件
	caKeyPath := filepath.Join(config.Dir, config.Filename+".key")
	caKeyFile, err := os.Create(caKeyPath)
	if err != nil {
		err = fmt.Errorf("创建 CA 私钥文件失败: %w", err)
		return
	}
	defer func() {
		_ = caKeyFile.Close()
	}()
	caPrivateKeyBytes, err := x509.MarshalPKCS8PrivateKey(caPrivateKey)
	if err != nil {
		err = fmt.Errorf("序列化 CA 私钥失败: %w", err)
		return
	}

	if err = pem.Encode(caKeyFile, &pem.Block{Type: "PRIVATE KEY", Bytes: caPrivateKeyBytes}); err != nil {
		err = fmt.Errorf("写入 CA 私钥失败: %w", err)
		return
	}

	return
}

type CertificateConfig struct {
	CADir       string
	CAFilename  string
	Common      *CAConfig
	DNSNames    []string
	IPAddresses []string
}

func NewDefaultCertificateConfig(dir, filename, caDir, caFilename, commonName string, validDays int64, dnsNames, ipAddresses []string) *CertificateConfig {
	config := &CertificateConfig{
		CADir:      caDir,
		CAFilename: caFilename,
		Common: &CAConfig{
			Dir:          dir,
			Filename:     filename,
			CommonName:   commonName,
			Organization: []string{"YM&MH Studio"},
			Country:      []string{"CN"},
			Province:     []string{"Liaoning"},
			Locality:     []string{"Shenyang"},
			ValidDays:    validDays,
			PrivateKeyConfig: &PrivateKeyConfig{
				KeyType: KeyTypeECDSA,
				Curve:   CurveP256,
			},
		},
	}
	if len(dnsNames) > 0 {
		config.DNSNames = dnsNames
	}
	if len(ipAddresses) > 0 {
		ipAddresses = append(ipAddresses, "127.0.0.1")
		ipAddresses = utils.RemoveDuplicates(ipAddresses)
	} else {
		ipAddresses = []string{"127.0.0.1"}
	}
	config.IPAddresses = ipAddresses

	return config
}

func LoadCA(dir string, fileName string) (caCert *x509.Certificate, caPrivateKey any, err error) {
	// 解析 CA 证书
	caCertPath := filepath.Join(dir, fileName+".crt")
	caCertPEM, err := os.ReadFile(caCertPath)
	if err != nil {
		err = fmt.Errorf("读取 CA 证书失败: %w", err)
		return
	}
	caCertBlock, _ := pem.Decode(caCertPEM)
	if caCertBlock == nil {
		err = fmt.Errorf("解析 CA 证书失败")
		return
	}
	caCert, err = x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		err = fmt.Errorf("解析 CA 证书失败: %w", err)
		return
	}

	// 解析 CA 私钥
	caKeyPath := filepath.Join(dir, fileName+".key")
	caKeyPEM, err := os.ReadFile(caKeyPath)
	if err != nil {
		err = fmt.Errorf("读取 CA 私钥失败: %w", err)
		return
	}
	caKeyBlock, _ := pem.Decode(caKeyPEM)
	if caKeyBlock == nil {
		err = fmt.Errorf("解析 CA 私钥失败")
		return
	}
	caPrivateKey, err = x509.ParsePKCS8PrivateKey(caKeyBlock.Bytes)
	if err != nil {
		err = fmt.Errorf("解析 CA 私钥失败: %w", err)
		return
	}
	// 验证私钥类型是否受支持
	switch caPrivateKey.(type) {
	case *rsa.PrivateKey, *ecdsa.PrivateKey, ed25519.PrivateKey:
		// 支持的密钥类型
	default:
		err = fmt.Errorf("不支持的CA私钥类型: %T", caPrivateKey)
	}

	return
}

func LoadOrCreateCA(config *CAConfig, isForce bool) (caCert *x509.Certificate, caPrivateKey any, err error) {
	caCertPath := filepath.Join(config.Dir, config.Filename+".crt")
	caKeyPath := filepath.Join(config.Dir, config.Filename+".key")

	_, errCaCert := os.Stat(caCertPath)
	_, errCaPrivateKey := os.Stat(caKeyPath)

	if errCaCert != nil || errCaPrivateKey != nil {
		if !os.IsNotExist(errCaCert) || !os.IsNotExist(errCaPrivateKey) {
			err = fmt.Errorf("errCaCert: %v, errCaPrivateKey: %v", errCaCert, errCaPrivateKey)
			return
		}
		if isForce {
			return GenerateCA(config)
		} else {
			err = errors.New("CA 文件不存在")
			return
		}
	}

	return LoadCA(config.Dir, config.Filename)
}

func GenerateCertificate(config *CertificateConfig) (cert *x509.Certificate, privateKey any, err error) {
	caCert, caPrivateKey, err := LoadCA(config.CADir, config.CAFilename)
	if err != nil {
		err = fmt.Errorf("加载 CA 失败: %w", err)
		return
	}

	if err = ensureDir(config.Common.Dir); err != nil {
		return
	}

	// 生成证书的公钥和私钥
	privateKey, err = generatePrivateKey(config.Common.PrivateKeyConfig)
	if err != nil {
		err = fmt.Errorf("生成私钥失败: %w", err)
		return
	}
	publicKey, err := getPublicKey(privateKey)
	if err != nil {
		return
	}

	// 生成证书
	serialNumber, err := generateSerialNumber()
	if err != nil {
		err = fmt.Errorf("生成序列号失败: %w", err)
		return
	}
	ipAddresses := make([]net.IP, 0, len(config.IPAddresses))
	for _, ip := range config.IPAddresses {
		ipAddresses = append(ipAddresses, net.ParseIP(ip))
	}
	cert = &x509.Certificate{
		SerialNumber: serialNumber,
		Issuer:       caCert.Subject,
		Subject: pkix.Name{
			CommonName:   config.Common.CommonName,
			Organization: config.Common.Organization,
			Country:      config.Common.Country,
			Province:     config.Common.Province,
			Locality:     config.Common.Locality,
		},
		DNSNames:    config.DNSNames,
		IPAddresses: ipAddresses,
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(time.Duration(config.Common.ValidDays) * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		BasicConstraintsValid: true,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, cert, caCert, publicKey, caPrivateKey)
	if err != nil {
		err = fmt.Errorf("创建证书失败: %w", err)
		return
	}
	certPath := filepath.Join(config.Common.Dir, config.Common.Filename+".crt")
	certFile, err := os.Create(certPath)
	if err != nil {
		err = fmt.Errorf("创建证书文件失败: %w", err)
		return
	}
	defer func() {
		_ = certFile.Close()
	}()
	if err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		err = fmt.Errorf("写入证书失败: %w", err)
		return
	}

	// 生成私钥文件
	keyPath := filepath.Join(config.Common.Dir, config.Common.Filename+".key")
	keyFile, err := os.Create(keyPath)
	if err != nil {
		err = fmt.Errorf("创建私钥文件失败: %w", err)
		return
	}
	defer func() {
		_ = keyFile.Close()
	}()
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		err = fmt.Errorf("序列化私钥失败: %w", err)
		return
	}

	if err = pem.Encode(keyFile, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBytes}); err != nil {
		err = fmt.Errorf("写入私钥失败: %w", err)
		return
	}

	return
}

func ValidateCert(certPath string) error {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("读取证书失败: %w", err)
	}

	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return fmt.Errorf("解析证书失败")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("解析证书失败: %w", err)
	}

	if time.Now().After(cert.NotAfter) {
		return fmt.Errorf("证书已过期")
	}

	if time.Now().Before(cert.NotBefore) {
		return fmt.Errorf("证书尚未生效")
	}

	return nil
}
