package pki

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"time"
)

// Opt represents a parameter for creating a new self-seigned
// certificate. See implementations below.
type Opt func(cert *x509.Certificate)

// SelfSignedCert will generate a self-signed certificate
// with all the passed options (see all Opt implementations).
//
// Return value is a *tls.Certificate, which will contain the
// X509 certificate struct:
//
// tlsCert := &tls.Certificate {
//		Certificate: [][]byte{x509Cert.Raw},
//		PrivateKey:  privateKey,
//		Leaf:        x509Cert,     // The x509 certificate struct.
// }
//
// The resultant *tls.Certificate can be used directly in a
// *tls.Config of the server.
func SelfSignedCert(opts ...Opt) (*tls.Certificate, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Go kit self-signed cert"},
			CommonName:   "127.0.0.1",
		},
		IPAddresses: []net.IP{net.ParseIP("0.0.0.0")},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(time.Hour * 24 * 365),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		PublicKey:   publicKey(privateKey),
	}
	for _, opt := range opts {
		opt(template)
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, publicKey(privateKey), privateKey)
	if err != nil {
		return nil, fmt.Errorf("pki: self-signed cert: %w", err)
	}
	x509Cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, fmt.Errorf("pki: self-signed cert: %w", err)
	}
	tlsCert := &tls.Certificate{
		Certificate: [][]byte{x509Cert.Raw},
		PrivateKey:  privateKey,
		Leaf:        x509Cert,
	}
	return tlsCert, nil
}

func WithCertSerialNumber(n int64) Opt {
	return func(crt *x509.Certificate) {
		crt.SerialNumber = big.NewInt(n)
	}
}

func WithCertCommonName(cn string) Opt {
	return func(crt *x509.Certificate) {
		crt.Subject.CommonName = cn
	}
}

func WithCertOrganization(org []string) Opt {
	return func(crt *x509.Certificate) {
		crt.Subject.Organization = org
	}
}

func WithCertDNSNames(dnsNames []string) Opt {
	return func(crt *x509.Certificate) {
		crt.DNSNames = dnsNames
	}
}

func WithCertIpAddresses(ips []string) Opt {
	return func(crt *x509.Certificate) {
		netIPs := make([]net.IP, len(ips))
		for i := 0; i < len(ips); i++ {
			netIPs[i] = net.ParseIP(ips[i])
		}
		crt.IPAddresses = netIPs
	}
}

func WithCertNotBefore(t time.Time) Opt {
	return func(crt *x509.Certificate) {
		crt.NotBefore = t
	}
}

func WithCertNotAfter(t time.Time) Opt {
	return func(crt *x509.Certificate) {
		crt.NotAfter = t
	}
}
