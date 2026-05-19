package monitor

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type CTClient struct {
	baseURL    string
	httpClient *http.Client
	maxRetries int
	retryDelay time.Duration
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
	return NewCTClientWithTimeout(baseURL, 30*time.Second)
}

func NewCTClientWithTimeout(baseURL string, timeout time.Duration) *CTClient {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &CTClient{
		baseURL:    normalizeCTBaseURL(baseURL),
		httpClient: &http.Client{Timeout: timeout},
		maxRetries: 2,
		retryDelay: 300 * time.Millisecond,
	}
}

func (c *CTClient) GetTreeSize(ctx context.Context) (int64, error) {
	var payload signedTreeHeadResponse
	if err := c.doJSONRequestWithRetry(ctx, c.baseURL+"/get-sth", &payload); err != nil {
		return 0, err
	}

	return payload.TreeSize, nil
}

func (c *CTClient) GetCertificates(ctx context.Context, start int64, end int64) ([]*x509.Certificate, error) {
	url := fmt.Sprintf("%s/get-entries?start=%d&end=%d", c.baseURL, start, end)
	var payload getEntriesResponse
	if err := c.doJSONRequestWithRetry(ctx, url, &payload); err != nil {
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

func (c *CTClient) doJSONRequestWithRetry(ctx context.Context, url string, out any) error {
	var lastErr error
	attempts := c.maxRetries + 1
	for attempt := 0; attempt < attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}

		response, err := c.httpClient.Do(request)
		if err != nil {
			lastErr = err
			if attempt < c.maxRetries {
				if waitErr := c.waitBeforeRetry(ctx, attempt); waitErr != nil {
					return waitErr
				}
				continue
			}
			return err
		}

		func() {
			defer response.Body.Close()
			if response.StatusCode < 200 || response.StatusCode >= 300 {
				lastErr = fmt.Errorf("ct request failed with status %d", response.StatusCode)
				return
			}
			lastErr = json.NewDecoder(response.Body).Decode(out)
		}()

		if lastErr == nil {
			return nil
		}

		if !shouldRetryStatus(response.StatusCode) || attempt >= c.maxRetries {
			return lastErr
		}

		if waitErr := c.waitBeforeRetry(ctx, attempt); waitErr != nil {
			return waitErr
		}
	}

	if lastErr == nil {
		lastErr = errors.New("ct request failed")
	}
	return lastErr
}

func shouldRetryStatus(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || statusCode >= 500
}

func (c *CTClient) waitBeforeRetry(ctx context.Context, attempt int) error {
	delay := c.retryDelay * time.Duration(attempt+1)
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
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
