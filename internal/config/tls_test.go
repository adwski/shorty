package config

import (
	"crypto/x509"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSelfSignedTLSConfig(t *testing.T) {
	tCfg, err := getSelfSignedTLSConfig("somehost")
	require.NoError(t, err)

	require.Len(t, tCfg.Certificates, 1)

	cert, err := x509.ParseCertificate(tCfg.Certificates[0].Certificate[0])
	require.NoError(t, err)

	require.Len(t, cert.DNSNames, 1)
	assert.Equal(t, "somehost", cert.DNSNames[0])
	assert.Equal(t, "somehost", cert.Subject.CommonName)

	assert.Len(t, cert.IPAddresses, 2)
	assert.Contains(t, cert.IPAddresses, net.IP{0x7f, 0x0, 0x0, 0x1})
	assert.Contains(t, cert.IPAddresses, net.IPv6loopback)
}
