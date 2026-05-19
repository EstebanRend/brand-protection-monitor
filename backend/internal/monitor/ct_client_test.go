package monitor

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestNormalizeCTBaseURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "keeps already valid base",
			input: "https://example.com/ct/v1",
			want:  "https://example.com/ct/v1",
		},
		{
			name:  "trims trailing slash",
			input: "https://example.com/ct/v1/",
			want:  "https://example.com/ct/v1",
		},
		{
			name:  "converts get roots endpoint into base",
			input: "https://example.com/ct/v1/get-roots",
			want:  "https://example.com/ct/v1",
		},
		{
			name:  "converts get sth endpoint into base",
			input: "https://example.com/ct/v1/get-sth",
			want:  "https://example.com/ct/v1",
		},
		{
			name:  "converts get entries endpoint into base",
			input: "https://example.com/ct/v1/get-entries",
			want:  "https://example.com/ct/v1",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := normalizeCTBaseURL(tt.input)
			if got != tt.want {
				t.Fatalf("normalizeCTBaseURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseCertificateFromExtraData(t *testing.T) {
	t.Parallel()

	certDER := mustCreateTestCertDER(t)
	directPayload := append(lengthPrefixed(certDER), []byte("ignored")...)
	chainPayload := append(lengthPrefixed(lengthPrefixed(certDER)), []byte("ignored")...)

	tests := []struct {
		name      string
		extraData string
		wantErr   bool
	}{
		{
			name:      "parses direct length prefixed cert",
			extraData: base64.StdEncoding.EncodeToString(directPayload),
		},
		{
			name:      "parses chain length prefixed cert",
			extraData: base64.StdEncoding.EncodeToString(chainPayload),
		},
		{
			name:      "rejects invalid payload",
			extraData: base64.StdEncoding.EncodeToString([]byte{0x01, 0x02}),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cert, err := parseCertificateFromExtraData(tt.extraData)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseCertificateFromExtraData returned error: %v", err)
			}
			if cert == nil {
				t.Fatalf("expected certificate, got nil")
			}
		})
	}
}

func TestShouldRetryStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
		want   bool
	}{
		{name: "retries too many requests", status: http.StatusTooManyRequests, want: true},
		{name: "retries internal server error", status: http.StatusInternalServerError, want: true},
		{name: "does not retry bad request", status: http.StatusBadRequest, want: false},
		{name: "does not retry not found", status: http.StatusNotFound, want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := shouldRetryStatus(tt.status)
			if got != tt.want {
				t.Fatalf("shouldRetryStatus(%d) = %t, want %t", tt.status, got, tt.want)
			}
		})
	}
}

func TestGetTreeSizeRetriesOnServerError(t *testing.T) {
	t.Parallel()

	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"tree_size": 123})
	}))
	defer server.Close()

	client := &CTClient{
		baseURL:    server.URL,
		httpClient: &http.Client{Timeout: 2 * time.Second},
		maxRetries: 2,
		retryDelay: 1 * time.Millisecond,
	}

	treeSize, err := client.GetTreeSize(t.Context())
	if err != nil {
		t.Fatalf("GetTreeSize returned error: %v", err)
	}
	if treeSize != 123 {
		t.Fatalf("GetTreeSize = %d, want 123", treeSize)
	}
	if atomic.LoadInt32(&attempts) != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
}

func mustCreateTestCertDER(t *testing.T) []byte {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "example.com",
		},
		DNSNames:              []string{"example.com", "www.example.com"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	return der
}

func lengthPrefixed(data []byte) []byte {
	prefix := []byte{
		byte(len(data) >> 16),
		byte(len(data) >> 8),
		byte(len(data)),
	}

	return append(prefix, data...)
}
