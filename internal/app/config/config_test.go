package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	testConfigPath = "/tmp/shorty-test-cfg.json"
	testCertPath   = "/tmp/shorty-test-cert.pem"
	testKeyPath    = "/tmp/shorty-test-key.pem"
)

var testConfig = `
{
  "storage": {
    "database_dsn": "postgres://qweasd.asd/db",
    "file_storage_path": "/qwe/qweasd",
    "trace_db": true
  },
  "tls": {
    "enable": true,
    "cert": "/tmp/shorty-test-cert.pem",
    "key": "/tmp/shorty-test-key.pem"
  },
  "listen_addr": "0.0.0.0:1111",
  "pprof_listen_addr": "0.0.0.0:2222",
  "base_url": "http://qwe.asd",
  "redirect_scheme": "http",
  "jwt_secret": "qweqwe",
  "trust_request_id": true
}
`

// Cert and Key are generated with:
// openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -sha256 -days 3650 -nodes \
// -subj "/C=XX/ST=State/L=City/O=Company/OU=Section/CN=CommonName"
// command.
var testKey = `-----BEGIN PRIVATE KEY-----
MIIJQgIBADANBgkqhkiG9w0BAQEFAASCCSwwggkoAgEAAoICAQDFdp7sdZCRG5Ka
ecIvyJaZm7MKqwlBTbFNJEJJdLsLp+jrrVKuSck6Tbvt/lhAZjc13J0J/Lhx3aN3
ywGD57PJJxUy0BkgNt6IMTTXLpfj2hHwza3jetVY1cZpIMoDrLn33EJ/Ri9hxdAS
/aQrrliyXLLU0FWmzPwLMDns1vJVM0VjrPcJyYh1IzcP+S1dwYGaY+i9xddp5peu
OSjbJy11lIVJIMJ/5F8cENyf/DSMzYGs8I54odgyZhIo9bQ0OFFkawCvdmkUUIXu
xXZc/JMZ16g8FCJDQTQTPJ51vfxMYGFIGcFMwBRcmMHH9V7+5UbufLJTTMEOs3YM
8sLB3gc3EHR+Jie8q3m0qzYal+zkLqVwm0xphLkm0S5wVlcf0is31LL0ySHbfAkB
cSxgkACwl0uwWXKIld7+M5of0d1n17cEkEtYz41GUcmcV8CePbFw1jD5RFKvWchw
nScgsSVaULFVEjGXkdyKz3pQrfSk2ms+mn0kFLfkXpRFRaYSlKDd5vzZvjGQZUh/
e+lopjEAXxHbMGsUiCnOs5Ee/KTXj0gt1/WR8wCqUg7sNxH/bb2FJKT8TxUmQ6Ph
oOBmAgcREpxLMTZXOKbnJxOg4v80JSPb74+CzOZmrJcplY2UeiIDRSGvSEt+A6/Y
NV9Y8LKSJIk0imeDrBH1XOnPCycNVwIDAQABAoICACp85x4+/6/RbH6To15vvUPp
FiG+ApxMENHl8uNmXBbadso7PZal5sgGUOEZQLj+pXOP+DRdbfyGMbXFdxqAQRWP
tMZ9s2JUnBZW7CU+78zFr+WOBBP16rEWMn6NYRpgUJWODbrgCbLygt6LOAd0GL6s
JoiXGU7uPW9U3anh6Du/7/bOEUvIUvXNcXwc5A+P4wiq5bnrt3mgddO1ld5t9CCg
J/u/skodg/+Ae1BrTo1bbMMe8bDwNhpGDzNEBxXTZmQCiB+5DUwjNWZWk6zXZfmC
Bz+CH4s0HPXkrb3s9rwiYtxGOmr+y4LPFacmW13iTJxlNWOqGMJGiCiqFZDsasWk
BQWcHoxmjUe+gjqF9Uj7VlY0ALtq+RLhRuZuXxDI0BAwu2UowiIp9OJdcys8T3EO
i/bbeikryMBZq150bRkxTmJoEuRmdmHYInBjxT9ldNvl82rf7Piems3+lwXaZWNe
/lscCW74igHHHsE4NE0STYh+3Ok6APoITKk64F6rW8dG8Nd+SOLC0IHdJMfK22Gy
gfbrBEC3DSgomoUwutcxlJXlTPvQG86cWvkmeiLd0EDr1uBTf3kI/+duEdKvZYbu
fVP+KOGy0PBtfOATt/oS5lbqsruTa245Rwj9Die+XpMW7z0Ok9EJ3jqfHPMOMl0f
BDJxbgM0Wpowuq7rrOKtAoIBAQDr6mtbXadIa0e2rRCyLnSQ77voN2tNFajV2OQR
pfk76zzeIwSDw9fuBCYlRy0QqyWb/zn4u0Css4vuFPmd4yxHok3tNDk9YHf6Ad36
sAedT50JGBoF9501BlO6Ufap0VE2/BT0TILA52ADtgHvHRqiDbTgomv9FLjjfxwJ
AXZIqVf5cDYlp0EJXbNLXLWn4C6EroaXpsICOE1+ReqrKJsJeh+9wyVoHNNFyMHG
kgkAZV8796SQkpMJXU4f9MjRLKC9X2xEbN+wJtoI/soON2ZGQUIK5yoM7uPdDHKj
I5BTwcm2cs+6aEuziXpkxn0t9D9rEd6RSF+HjVwTlHhjvrn7AoIBAQDWRipodsPT
7Rkv3QcrwjaQiIBdO4ImBS6WaRP0yWuAiAQGlz0NCfGUoE3oAsenQsh18DR9EoL5
5m5er9F0UItOcYEVT3uyVHLyHwmCNKFcZkTHDiDDukUnvKWK/mb5MOATBe85XHi6
UebQzdLA9N83POjfndjHw8YhaoMUsW+BpkWj6giz6zmNZn8Tzx37ar4oa6MmXaEm
9FLpFrTalaBV2IV3s6uGyPV695hBJhWiivBhPf3Y7Ywqi9gsBKTWu4QzyxAmZrq3
ygQDuai30t90g0b96EDR687dD9xspvHsOPejwQld0VcCHBBvlI/bFJqSTU9RPhBy
qB3Qfi63rVdVAoIBAQCnRqdquEQ12EOYJqyQ+tmSLzo4lJsTpEj7oHdOgDXxo4Gc
LI4187Z2wSBfDFHK4N+g9d6gG/3mtsSAQwUfS4YJIO9KQt5XQ8CuV4aTCx/LDjlV
ym4hTwp8H4lcsmNI0+9XInSlKF8J8sUkvHgwmJ1azIc1RFV3tKrIPEefpxa8vL5x
UnxCRI6b2oGX7Rus9gtl7u3mN8qWkl94KpETXY5YsOwyvF0Yrl+ruoaTitaxHi/h
sF1SWWvClxMfG40MrC0pObDl76DIITQ2bprMa8GsDPeMDY7GbtjI0tuyCzR5/w1M
vanHDc6nddKABDGcVPRmsdvzfbKmxbfE9mBKsWDjAoIBADBC44Bd56OPDpI6PUg8
2R9ar1bQdXLszd5w0l7bEwyDFi0J2WVbbP8l0AZGwgNJRm9R5/CXv8pbfVZ0UT/a
eFX1uLY9DcZPwQgJt5GNGx79pdYTt4t+I99cXQjeXgEYYg/G0WfhDQwiMFOtWB+/
x5rgbC8ZlV8BGFokbFu3zz6rXZoat1UW0QKpKEwk5ULgeu4NuFSG2Co284muZJOr
Cc7sEruwSxzznF7S74cU9eCDXLr6RHecoWpfzxOzBBGdcJxdy0hq8Pk+VgMkyPyG
UXAjtVSxABCIBTLDRMlwHKMw/Y3zh5GK+gbunUEUfAZDMMFtCkVpLjk05mo0UX9r
WnUCggEAd90VJLwILXfKwLMOj27G5CfM+oL/FJBdY/ekjhej3OVgyrU/zqs5EHIr
t31b52BjA/HdgaQnUSGU7iXUD7eO1UP6iVyGgrwqFUxWk/UfQ04dVA9XpluU3NeW
4XsFM8OU2fd1K1Trd4T5QnyotKkjEpZYSJtgiy2+raDeSawvdzyh8V4XDj8vdcMe
5yFc+AdfhD1CwA7zetoUhbIcy7gFOzKxGeWvIGHCuMSodxZyWWavLcVS3M02kOPg
2LvXZ93lZr+WK95n6fc8Uk3oCcZW5va6biPGdb+Eqj43e8UNFFzAaBgbo3QVFV5p
41+ugQjYpkzPzw8vG+k3tOqo5kSvWQ==
-----END PRIVATE KEY-----
`
var testCert = `-----BEGIN CERTIFICATE-----
MIIFqzCCA5OgAwIBAgIUHd0CAXcSFQORYS+vXffvnVUfkFMwDQYJKoZIhvcNAQEL
BQAwZTELMAkGA1UEBhMCWFgxDjAMBgNVBAgMBVN0YXRlMQ0wCwYDVQQHDARDaXR5
MRAwDgYDVQQKDAdDb21wYW55MRAwDgYDVQQLDAdTZWN0aW9uMRMwEQYDVQQDDApD
b21tb25OYW1lMB4XDTI0MDMyNjIwNTUyMloXDTM0MDMyNDIwNTUyMlowZTELMAkG
A1UEBhMCWFgxDjAMBgNVBAgMBVN0YXRlMQ0wCwYDVQQHDARDaXR5MRAwDgYDVQQK
DAdDb21wYW55MRAwDgYDVQQLDAdTZWN0aW9uMRMwEQYDVQQDDApDb21tb25OYW1l
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAxXae7HWQkRuSmnnCL8iW
mZuzCqsJQU2xTSRCSXS7C6fo661SrknJOk277f5YQGY3NdydCfy4cd2jd8sBg+ez
yScVMtAZIDbeiDE01y6X49oR8M2t43rVWNXGaSDKA6y599xCf0YvYcXQEv2kK65Y
slyy1NBVpsz8CzA57NbyVTNFY6z3CcmIdSM3D/ktXcGBmmPovcXXaeaXrjko2yct
dZSFSSDCf+RfHBDcn/w0jM2BrPCOeKHYMmYSKPW0NDhRZGsAr3ZpFFCF7sV2XPyT
GdeoPBQiQ0E0Ezyedb38TGBhSBnBTMAUXJjBx/Ve/uVG7nyyU0zBDrN2DPLCwd4H
NxB0fiYnvKt5tKs2Gpfs5C6lcJtMaYS5JtEucFZXH9IrN9Sy9Mkh23wJAXEsYJAA
sJdLsFlyiJXe/jOaH9HdZ9e3BJBLWM+NRlHJnFfAnj2xcNYw+URSr1nIcJ0nILEl
WlCxVRIxl5Hcis96UK30pNprPpp9JBS35F6URUWmEpSg3eb82b4xkGVIf3vpaKYx
AF8R2zBrFIgpzrORHvyk149ILdf1kfMAqlIO7DcR/229hSSk/E8VJkOj4aDgZgIH
ERKcSzE2Vzim5ycToOL/NCUj2++PgszmZqyXKZWNlHoiA0Uhr0hLfgOv2DVfWPCy
kiSJNIpng6wR9VzpzwsnDVcCAwEAAaNTMFEwHQYDVR0OBBYEFAFX6XWjnmRU8HTh
VZ+OtRZmS/tQMB8GA1UdIwQYMBaAFAFX6XWjnmRU8HThVZ+OtRZmS/tQMA8GA1Ud
EwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggIBAMQaRNld4lfSk645NIgskNCj
3cSQHsTLogbicPMcC6FsAVWokipvDQmjTWKLlskBFmotX2M8nkBFoGQFda29Wxv6
kFaNRu3gPjFODrG73IrUYD7CYnTFfsAvUd5G+ZJNeHZXKDeM0250YSFyti0t+z6f
VXWwKu7hGc4w60QmSrYFyf1ikjbOvkiKzB/CIK/PrnqfLW2KQzoTTdjjxnA2spwk
82NeZXeVTSVxfbDPxpw148wEt5Z67yHKFfzzRY+oNn/m48VljnsG3nioXjk8vSRk
83i7kCzpa/YVvTZ4xlq7PFGu4LIj8y2ubuUQRdcv96PkTdxGp/bg0SxL0EhonIWr
5ayVDzpTeE3Vh5s9AS2sc1e+0JQvy7+wxa2CacMlKgouRntU0+h9LrESP55sMElq
fbk0G9civ5kewMXRjTXTOpB1QUwEHC5oSlLNlcSpKQwziQPVDzti7thOs4Nqrr8+
glXlpj7EFKIqNk4sHtkOE3eGMqg2fpC5jR4Z/zmu73hj9gl9L3YoFRDcc/FA5saj
Knkvk1DSkMG+/9455DpYLpjEDRHszzg6tnZi/3QEwBgOAy8LZ6vzfjNapRyImPd2
bOj8jarTw4hkL3zm6G4wqYhvY/SExEs04uYQ4jQKUP2qJoUt6xJpbSit32ayylxp
zpSraVpUe1NHT2OlJQOc
-----END CERTIFICATE-----
`

func TestNew(t *testing.T) {
	err := os.WriteFile(testConfigPath, []byte(testConfig), os.ModePerm)
	require.NoError(t, err)
	err = os.WriteFile(testKeyPath, []byte(testKey), os.ModePerm)
	require.NoError(t, err)
	err = os.WriteFile(testCertPath, []byte(testCert), os.ModePerm)
	require.NoError(t, err)

	t.Setenv("CONFIG", testConfigPath)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	cfg, err := New(logger)
	require.NoError(t, err)

	assert.NotNil(t, cfg.GetTLSConfig())
	assert.NotNil(t, cfg.Storage)

	assert.Equal(t, "/qwe/qweasd", cfg.Storage.FileStoragePath)
	assert.Equal(t, "postgres://qweasd.asd/db", cfg.Storage.DatabaseDSN)
	assert.True(t, cfg.Storage.TraceDB)

	assert.True(t, cfg.TLS.Enable)
	assert.Equal(t, testCertPath, cfg.TLS.CertPath)
	assert.Equal(t, testKeyPath, cfg.TLS.KeyPath)

	assert.Equal(t, "0.0.0.0:1111", cfg.ListenAddr)
	assert.Equal(t, "0.0.0.0:2222", cfg.PprofServerAddr)
	assert.Equal(t, "http://qwe.asd", cfg.BaseURL)
	assert.Equal(t, "qwe.asd", cfg.ServedHost)
	assert.Equal(t, "http", cfg.ServedScheme)
	assert.Equal(t, "qweqwe", cfg.JWTSecret)
	assert.True(t, cfg.TrustRequestID)

	_ = os.Remove(testConfigPath)
	_ = os.Remove(testKeyPath)
	_ = os.Remove(testCertPath)
}
