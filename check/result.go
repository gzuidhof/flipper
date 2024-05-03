package check

import (
	"crypto/tls"
	"fmt"
	"time"
)

// Result is the result of a health check.
type Result interface {
	// Healthy returns true if the check is healthy.
	Healthy() bool
}

// HTTPCheckResult is the result of a HTTP health check.
type HTTPCheckResult struct {
	// Error is the error that occurred during the check, or nil if the check was successful.
	Error error

	// StatusCode is the HTTP status code of the response.
	StatusCode int

	// Duration is the time it took to perform the check.
	TLS TLSCertificatesResult
}

// Healthy returns true if the check is healthy.
func (r HTTPCheckResult) Healthy() bool {
	return r.Error == nil
}

// Errorf sets the error of the result, it's a convenience method to set the error with a formatted string.
func (r HTTPCheckResult) Errorf(fmtString string, args ...any) HTTPCheckResult {
	r.Error = fmt.Errorf(fmtString, args...)
	return r
}

var _ Result = (*HTTPCheckResult)(nil)

// TLSCertificatesResult represents the result of a TLS certificate check.
// It only considers the leaf certificate - which is generally the one that matters.
type TLSCertificatesResult struct {
	// IsTLS is true if the connection is using TLS.
	IsTLS bool
	// NotAfter is the expiration date of the certificate.
	NotAfter time.Time
	// NotBefore is the start date of the certificate.
	NotBefore time.Time

	// DNSNames is a list of DNS names in the certificate.
	DNSNames []string
}

// TLSCertificatesResultFromConnectionState creates a TLSCeriticatesResult from a tls.ConnectionState.
func TLSCertificatesResultFromConnectionState(s *tls.ConnectionState) TLSCertificatesResult {
	r := TLSCertificatesResult{
		IsTLS: false,
	}
	if s == nil || len(s.PeerCertificates) == 0 {
		return r
	}
	r.IsTLS = true

	// The first certificate is the leaf certificate.
	pc := s.PeerCertificates[0]

	r.NotAfter = pc.NotAfter
	r.NotBefore = pc.NotBefore
	r.DNSNames = append([]string(nil), pc.DNSNames...)

	return r
}

// ExpiresWithin returns true if the certificate expires within the given duration.
func (r TLSCertificatesResult) ExpiresWithin(d time.Duration) bool {
	return time.Now().Add(d).After(r.NotAfter)
}
