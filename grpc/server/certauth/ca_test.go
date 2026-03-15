package certauth

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCA_InitAndLoad(t *testing.T) {
	dir := t.TempDir()

	ca1, err := LoadOrCreateCA(dir)
	require.NoError(t, err)
	require.NotNil(t, ca1)

	ca2, err := LoadOrCreateCA(dir)
	require.NoError(t, err)
	require.NotNil(t, ca2)

	assert.Equal(t, ca1.Cert.SerialNumber, ca2.Cert.SerialNumber,
		"reloaded CA must have the same serial number")
	assert.Equal(t, "jniservice CA", ca2.Cert.Subject.CommonName)
	assert.True(t, ca2.Cert.IsCA)
}

func TestCA_SignCSR(t *testing.T) {
	dir := t.TempDir()

	ca, err := LoadOrCreateCA(dir)
	require.NoError(t, err)

	csrPEM, _, err := GenerateCSR("test-client")
	require.NoError(t, err)

	certPEM, err := ca.SignCSR(csrPEM)
	require.NoError(t, err)

	cert, err := ParseCertPEM(certPEM)
	require.NoError(t, err)

	assert.Equal(t, "test-client", cert.Subject.CommonName)

	// Verify certificate chains to the CA.
	pool := x509.NewCertPool()
	pool.AddCert(ca.Cert)
	_, err = cert.Verify(x509.VerifyOptions{
		Roots:     pool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	})
	require.NoError(t, err, "client cert must verify against CA")

	// Verify ExtKeyUsage includes ClientAuth.
	require.NotEmpty(t, cert.ExtKeyUsage)
	assert.Contains(t, cert.ExtKeyUsage, x509.ExtKeyUsageClientAuth)

	// Negative: cert must NOT verify with an empty pool.
	emptyPool := x509.NewCertPool()
	_, err = cert.Verify(x509.VerifyOptions{
		Roots:     emptyPool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	})
	assert.Error(t, err, "client cert must not verify against empty pool")
}

func TestCA_SignCSR_InvalidCSR(t *testing.T) {
	dir := t.TempDir()

	ca, err := LoadOrCreateCA(dir)
	require.NoError(t, err)

	_, err = ca.SignCSR([]byte("not a valid CSR"))
	assert.Error(t, err)
}

func TestCA_GenerateCSR(t *testing.T) {
	csrPEM, keyPEM, err := GenerateCSR("my-device")
	require.NoError(t, err)
	require.NotEmpty(t, csrPEM)
	require.NotEmpty(t, keyPEM)

	// Parse and verify CSR has the correct CN.
	csr, err := parseCSRPEM(csrPEM)
	require.NoError(t, err)
	assert.Equal(t, "my-device", csr.Subject.CommonName)

	// Verify CSR signature is valid.
	assert.NoError(t, csr.CheckSignature())
}

func TestCA_CertPEM(t *testing.T) {
	dir := t.TempDir()

	ca, err := LoadOrCreateCA(dir)
	require.NoError(t, err)

	pemData := ca.CertPEM()
	require.NotEmpty(t, pemData)

	cert, err := ParseCertPEM(pemData)
	require.NoError(t, err)
	assert.Equal(t, ca.Cert.SerialNumber, cert.SerialNumber)
}

func TestParseCertPEM_Invalid(t *testing.T) {
	_, err := ParseCertPEM([]byte("not a certificate"))
	assert.Error(t, err)
}

func parseCSRPEM(pemData []byte) (*x509.CertificateRequest, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode CSR PEM")
	}
	return x509.ParseCertificateRequest(block.Bytes)
}
