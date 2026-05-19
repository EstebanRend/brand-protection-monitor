package monitor

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type CTClient struct {
	baseURL    string
	httpClient *http.Client
}

type signedTreeHeadResponse struct {
	TreeSize int64 `json:"tree_size"`
}

type getEntriesResponse struct {
	Entries []ctEntry `json:"entries"`
}

type ctEntry struct {
	LeafInput string `json:"leaf_input"`
	ExtraData string `json:"extra_data"`
}

func NewCTClient(baseURL string) *CTClient {
	return &CTClient{
		baseURL:    normalizeCTBaseURL(baseURL),
		httpClient: http.DefaultClient,
	}
}

func (c *CTClient) GetTreeSize(ctx context.Context) (int64, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/get-sth", nil)
	if err != nil {
		return 0, err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return 0, fmt.Errorf("get-sth failed with status %d", response.StatusCode)
	}

	var payload signedTreeHeadResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return 0, err
	}

	return payload.TreeSize, nil
}

func (c *CTClient) GetCertificates(ctx context.Context, start int64, end int64) ([]*x509.Certificate, error) {
	url := fmt.Sprintf("%s/get-entries?start=%d&end=%d", c.baseURL, start, end)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("get-entries failed with status %d", response.StatusCode)
	}

	var payload getEntriesResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}

	certificates := make([]*x509.Certificate, 0)
	for _, entry := range payload.Entries {
		parsed, err := parseCertificateFromExtraData(entry.ExtraData)
		if err == nil && parsed != nil {
			certificates = append(certificates, parsed)
		}
	}

	return certificates, nil
}

func parseCertificateFromExtraData(extraData string) (*x509.Certificate, error) {
	rawBytes, err := base64.StdEncoding.DecodeString(extraData)
	if err != nil {
		return nil, err
	}

	// Some CT logs return extra_data as:
	// 1) cert_length(3) + cert_der + ...
	// 2) chain_length(3) + [cert_length(3) + cert_der + ...]
	// Try direct first, then with an outer chain-length prefix.
	if cert, err := parseFirstLengthPrefixedCertificate(rawBytes); err == nil {
		return cert, nil
	}

	if len(rawBytes) >= 6 {
		chainLength := readUint24(rawBytes[:3])
		if chainLength > 0 && len(rawBytes) >= 3+chainLength {
			if cert, err := parseFirstLengthPrefixedCertificate(rawBytes[3 : 3+chainLength]); err == nil {
				return cert, nil
			}
		}
	}

	return nil, fmt.Errorf("failed to parse certificate from extra_data")
}

func parseFirstLengthPrefixedCertificate(payload []byte) (*x509.Certificate, error) {
	if len(payload) < 3 {
		return nil, fmt.Errorf("payload too short")
	}

	certLength := readUint24(payload[:3])
	if certLength <= 0 || len(payload) < 3+certLength {
		return nil, fmt.Errorf("invalid certificate length")
	}

	certDER := payload[3 : 3+certLength]
	return x509.ParseCertificate(certDER)
}

func readUint24(data []byte) int {
	return int(data[0])<<16 | int(data[1])<<8 | int(data[2])
}

func normalizeCTBaseURL(baseURL string) string {
	normalized := strings.TrimSpace(baseURL)
	normalized = strings.TrimRight(normalized, "/")

	for _, suffix := range []string{"/get-roots", "/get-sth", "/get-entries"} {
		if strings.HasSuffix(normalized, suffix) {
			normalized = strings.TrimSuffix(normalized, suffix)
			break
		}
	}

	return normalized
}
