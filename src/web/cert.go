package web

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"v2rayn-go/config"
)

// ensureCerts 确保证书文件存在，不存在则自动生成自签名证书
// 证书放在 certDir 目录下（默认为应用运行目录/certs）
// 返回证书文件和私钥文件的绝对路径
func ensureCerts(certDir string) (certFile, keyFile string, err error) {
	certFile = filepath.Join(certDir, "server.crt")
	keyFile = filepath.Join(certDir, "server.key")

	// 如果证书和私钥都已存在，直接返回
	if fileExists(certFile) && fileExists(keyFile) {
		return certFile, keyFile, nil
	}

	slog.Info("generating self-signed TLS certificate", "dir", certDir)

	// 创建证书目录
	if err := os.MkdirAll(certDir, 0700); err != nil {
		return "", "", fmt.Errorf("create cert dir: %w", err)
	}

	// 生成 ECDSA P-256 私钥（比 RSA 更快、更安全）
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("generate private key: %w", err)
	}

	// 生成随机序列号
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return "", "", fmt.Errorf("generate serial number: %w", err)
	}

	// 证书模板
	// 有效期 10 年，覆盖 localhost 常见地址
	notBefore := time.Now()
	notAfter := notBefore.Add(10 * 365 * 24 * time.Hour)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"v2rayN-Go"},
			CommonName:   "v2rayN-Go Self-Signed",
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},

		// SAN（Subject Alternative Names）：覆盖常见本地地址
		IPAddresses: []net.IP{
			net.ParseIP("127.0.0.1"),
			net.ParseIP("::1"),
			net.ParseIP("0.0.0.0"),
		},
		DNSNames: []string{"localhost"},
	}

	// 自签名：用自身私钥签发证书
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", fmt.Errorf("create certificate: %w", err)
	}

	// 写入证书文件（PEM 编码，使用 AtomicWriteFile 保证断电安全）
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err := config.AtomicWriteFile(certFile, certPEM, 0644); err != nil {
		return "", "", fmt.Errorf("write cert file: %w", err)
	}

	// 写入私钥文件（权限 0600，仅 owner 可读写）
	keyDER, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return "", "", fmt.Errorf("marshal private key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	if err := config.AtomicWriteFile(keyFile, keyPEM, 0600); err != nil {
		return "", "", fmt.Errorf("write key file: %w", err)
	}

	slog.Info("self-signed TLS certificate generated", "cert", certFile, "key", keyFile, "expires", notAfter.Format("2006-01-02"))
	return certFile, keyFile, nil
}

// startHTTPS 启动 HTTPS 服务器
// 如果证书不存在则自动生成自签名证书
func (s *Server) startHTTPS(handler http.Handler, certDir string) error {
	// ensureCerts 会自动检测并生成证书
	certFile, keyFile, err := ensureCerts(certDir)
	if err != nil {
		return fmt.Errorf("ensure certs: %w", err)
	}

	addr := s.cfg.GetListenAddr()
	slog.Info("HTTPS server starting", "addr", addr, "cert", certFile)
	return http.ListenAndServeTLS(addr, certFile, keyFile, handler)
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
