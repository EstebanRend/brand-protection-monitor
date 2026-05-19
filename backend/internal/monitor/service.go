package monitor

import (
	"context"
	"crypto/x509"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"brand-protection-monitor/backend/internal/certificates"
	"brand-protection-monitor/backend/internal/keywords"
)

var ErrRunAlreadyInProgress = errors.New("monitor run already in progress")

type Service struct {
	ctClient        *CTClient
	keywordRepo     *keywords.Repository
	certificateRepo *certificates.Repository
	stateRepo       *StateRepository
	batchSize       int64
	sourceLog       string
	runMu           sync.Mutex
	isRunning       bool
}

func NewService(
	ctClient *CTClient,
	keywordRepo *keywords.Repository,
	certificateRepo *certificates.Repository,
	stateRepo *StateRepository,
	batchSize int64,
	sourceLog string,
) *Service {
	return &Service{
		ctClient:        ctClient,
		keywordRepo:     keywordRepo,
		certificateRepo: certificateRepo,
		stateRepo:       stateRepo,
		batchSize:       batchSize,
		sourceLog:       sourceLog,
	}
}

func (s *Service) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	if err := s.RunOnce(ctx); err != nil && !errors.Is(err, ErrRunAlreadyInProgress) {
		log.Printf("initial monitor run failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.RunOnce(ctx); err != nil && !errors.Is(err, ErrRunAlreadyInProgress) {
				log.Printf("monitor run failed: %v", err)
			}
		}
	}
}

func (s *Service) RunOnce(ctx context.Context) error {
	if !s.tryStartRun() {
		return ErrRunAlreadyInProgress
	}
	defer s.finishRun()

	currentState, err := s.stateRepo.Get(ctx)
	if err != nil {
		return err
	}

	treeSize, err := s.ctClient.GetTreeSize(ctx)
	if err != nil {
		_ = s.stateRepo.Update(ctx, currentState.LastTreeSize, 0, "error")
		return err
	}

	start, end, nextTreeSize, hasWork := computeProcessingRange(currentState.LastTreeSize, treeSize, s.batchSize)
	if !hasWork {
		return s.stateRepo.Update(ctx, nextTreeSize, 0, "active")
	}

	certs, err := s.ctClient.GetCertificates(ctx, start, end)
	if err != nil {
		_ = s.stateRepo.Update(ctx, currentState.LastTreeSize, 0, "error")
		return err
	}

	monitoredKeywords, err := s.keywordRepo.List(ctx)
	if err != nil {
		_ = s.stateRepo.Update(ctx, currentState.LastTreeSize, len(certs), "error")
		return err
	}

	saveFailures := 0
	for _, cert := range certs {
		saveFailures += s.matchAndStore(ctx, cert, monitoredKeywords)
	}

	if saveFailures > 0 {
		log.Printf("monitor cycle completed with %d save failures", saveFailures)
	}

	return s.stateRepo.Update(ctx, nextTreeSize, len(certs), "active")
}

func (s *Service) matchAndStore(ctx context.Context, cert *x509.Certificate, monitoredKeywords []keywords.Keyword) int {
	domains := extractDomains(cert)
	issuer := cert.Issuer.CommonName
	if issuer == "" {
		issuer = cert.Issuer.String()
	}

	failures := 0
	for _, domain := range domains {
		normalizedDomain := strings.ToLower(domain)

		for _, keyword := range monitoredKeywords {
			normalizedKeyword := strings.ToLower(strings.TrimSpace(keyword.Value))
			if normalizedKeyword == "" {
				continue
			}

			if strings.Contains(normalizedDomain, normalizedKeyword) {
				if err := s.certificateRepo.SaveMatch(ctx, certificates.NewMatchedCertificate{
					Domain:         domain,
					Issuer:         issuer,
					NotBefore:      cert.NotBefore,
					NotAfter:       cert.NotAfter,
					MatchedKeyword: keyword.Value,
					SourceLog:      s.sourceLog,
				}); err != nil {
					log.Printf("failed to save match (domain=%s, keyword=%s): %v", domain, keyword.Value, err)
					failures++
				}
			}
		}
	}

	return failures
}

func extractDomains(cert *x509.Certificate) []string {
	seen := map[string]bool{}
	domains := make([]string, 0)

	addDomain := func(value string) {
		normalized := strings.ToLower(strings.TrimSpace(value))
		if normalized == "" || seen[normalized] {
			return
		}
		seen[normalized] = true
		domains = append(domains, normalized)
	}

	addDomain(cert.Subject.CommonName)
	for _, name := range cert.DNSNames {
		addDomain(name)
	}

	return domains
}

func (s *Service) tryStartRun() bool {
	s.runMu.Lock()
	defer s.runMu.Unlock()

	if s.isRunning {
		return false
	}

	s.isRunning = true
	return true
}

func (s *Service) finishRun() {
	s.runMu.Lock()
	defer s.runMu.Unlock()
	s.isRunning = false
}

func computeProcessingRange(previousTreeSize int64, currentTreeSize int64, batchSize int64) (start int64, end int64, nextTreeSize int64, hasWork bool) {
	if currentTreeSize <= 0 {
		return 0, -1, 0, false
	}

	if batchSize <= 0 {
		batchSize = 1
	}

	latestIndex := currentTreeSize - 1
	if previousTreeSize <= 0 {
		start = currentTreeSize - batchSize
		if start < 0 {
			start = 0
		}
	} else {
		start = previousTreeSize
	}

	if start > latestIndex {
		return 0, -1, currentTreeSize, false
	}

	end = latestIndex
	maxEnd := start + batchSize - 1
	if maxEnd < end {
		end = maxEnd
	}

	return start, end, end + 1, true
}
