package pki_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/kit/pki"
)

func TestSelfSignedCert(t *testing.T) {
	var sn int64 = 2
	notBefore := time.Now()
	notAfter := notBefore.Add(10 * time.Hour)
	IPAddresses := []string{"192.168.1.2", "192.168.1.3"}
	DNSNames := []string{"example.com"}
	org := []string{"testing"}
	cn := "example.com"

	tlsCrt, err := pki.SelfSignedCert(
		pki.WithCertSerialNumber(sn),
		pki.WithCertCommonName(cn),
		pki.WithCertOrganization(org),
		pki.WithCertIpAddresses(IPAddresses),
		pki.WithCertDNSNames(DNSNames),
		pki.WithCertNotBefore(notBefore),
		pki.WithCertNotAfter(notAfter),
	)
	require.NoError(t, err)

	require.NotEmpty(t, tlsCrt.Certificate, "raw certificate is not being passed to TLS cert struct")
	require.NotEmpty(t, tlsCrt.PrivateKey, "private key is not being passed to TLS cert struct")
	require.NotEmpty(t, tlsCrt.Leaf, "x509 certificate is not being passed to TLS cert struct")

	x509Crt := tlsCrt.Leaf

	assert.Equal(t, big.NewInt(sn), x509Crt.SerialNumber)
	assert.Equal(t, cn, x509Crt.Subject.CommonName)
	assert.Equal(t, org, x509Crt.Subject.Organization)
	require.Len(t, x509Crt.IPAddresses, 2)
	assert.Equal(t, IPAddresses, []string{x509Crt.IPAddresses[0].String(), x509Crt.IPAddresses[1].String()})
	require.Len(t, x509Crt.DNSNames, 1)
	assert.Equal(t, x509Crt.DNSNames, DNSNames)

	timeLayout := "20060102150405" // to second precision
	assert.Equal(t, notBefore.UTC().Format(timeLayout), x509Crt.NotBefore.Format(timeLayout))
	assert.Equal(t, notAfter.UTC().Format(timeLayout), x509Crt.NotAfter.Format(timeLayout))
}
