package config

import (
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"time"

	"go.uber.org/zap"
)

const (
	caOrg        = "Shorty"
	caCountry    = "RU"
	caValidYears = 10

	privateKeyRSALen = 4096

	minTLSVersion = tls.VersionTLS13
)

var (
	caSubjectKeyIdentifier = []byte{1, 2, 3, 4, 6}
)

func getTLSConfig(logger *zap.Logger, cfg *TLS, host string) (*tls.Config, error) {
	if cfg.UseSelfSigned {
		// generate self signed
		logger.Warn("generating self-signed cert")
		logger.Warn("self signed cert will have loopback ip addresses and served host as SANs")
		return getSelfSignedTLSConfig(host)
	}

	if cfg.KeyPath != "" && cfg.CertPath != "" {
		// use provided key and cert
		cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("cannot load x509 cert and key: %w", err)
		}

		return &tls.Config{
			MinVersion:   minTLSVersion,
			Certificates: []tls.Certificate{cert},
		}, nil
	}

	var err error
	if cfg.KeyPath == "" {
		err = errors.Join(err, fmt.Errorf("key path is empty"))
	}
	if cfg.CertPath == "" {
		err = errors.Join(err, fmt.Errorf("cert path is empty"))
	}
	return nil, err
}

func getSelfSignedTLSConfig(cn string) (*tls.Config, error) {
	key, err := rsa.GenerateKey(crand.Reader, privateKeyRSALen)
	if err != nil {
		return nil, fmt.Errorf("cannot generate rsa private key: %w", err)
	}

	ca := getCA(cn)
	cert, err := x509.CreateCertificate(crand.Reader, ca, ca, &key.PublicKey, key)
	if err != nil {
		return nil, fmt.Errorf("cannot create x509 certificate: %w", err)
	}

	return &tls.Config{
		MinVersion: minTLSVersion,
		Certificates: []tls.Certificate{{
			Certificate: [][]byte{cert},
			PrivateKey:  key,
		}}}, nil
}

func getCA(cn string) *x509.Certificate {
	return &x509.Certificate{
		SerialNumber: big.NewInt(rand.Int63()),
		Subject: pkix.Name{
			Country:      []string{caCountry},
			Organization: []string{caOrg},
			CommonName:   cn,
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}, //nolint:gomnd // ip addr
		DNSNames:     []string{cn},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(caValidYears, 0, 0),
		SubjectKeyId: caSubjectKeyIdentifier,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
}
