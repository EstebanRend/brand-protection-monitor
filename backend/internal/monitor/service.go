package monitor

import (
    "context"
    "crypto/x509"
    "log"
    "strings"
    "time"

    "brand-protection-monitor/backend/internal/certificates"
    "brand-protection-monitor/backend/internal/keywords"
)

type Service struct {
    ctClient        *CTClient
    keywordRepo     *keywords.Repository
    certificateRepo *certificates.Repository
    stateRepo       *StateRepository
    batchSize       int64
    sourceLog       string
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

    if err := s.RunOnce(ctx); err != nil {
        log.Printf("initial monitor run failed: %v", err)
    }

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := s.RunOnce(ctx); err != nil {
                log.Printf("monitor run failed: %v", err)
            }
        }
    }
}

func (s *Service) RunOnce(ctx context.Context) error {
    treeSize, err := s.ctClient.GetTreeSize(ctx)
    if err != nil {
        _ = s.stateRepo.Update(ctx, 0, 0, "error")
        return err
    }

    if treeSize <= 0 {
        return s.stateRepo.Update(ctx, treeSize, 0, "active")
    }

    start := treeSize - s.batchSize
    if start < 0 {
        start = 0
    }
    end := treeSize - 1

    certs, err := s.ctClient.GetCertificates(ctx, start, end)
    if err != nil {
        _ = s.stateRepo.Update(ctx, treeSize, 0, "error")
        return err
    }

    monitoredKeywords, err := s.keywordRepo.List(ctx)
    if err != nil {
        _ = s.stateRepo.Update(ctx, treeSize, len(certs), "error")
        return err
    }

    for _, cert := range certs {
        s.matchAndStore(ctx, cert, monitoredKeywords)
    }

    return s.stateRepo.Update(ctx, treeSize, len(certs), "active")
}

func (s *Service) matchAndStore(ctx context.Context, cert *x509.Certificate, monitoredKeywords []keywords.Keyword) {
    domains := extractDomains(cert)
    issuer := cert.Issuer.CommonName
    if issuer == "" {
        issuer = cert.Issuer.String()
    }

    for _, domain := range domains {
        normalizedDomain := strings.ToLower(domain)

        for _, keyword := range monitoredKeywords {
            normalizedKeyword := strings.ToLower(strings.TrimSpace(keyword.Value))
            if normalizedKeyword == "" {
                continue
            }

            if strings.Contains(normalizedDomain, normalizedKeyword) {
                _ = s.certificateRepo.SaveMatch(ctx, certificates.NewMatchedCertificate{
                    Domain:         domain,
                    Issuer:         issuer,
                    NotBefore:      cert.NotBefore,
                    NotAfter:       cert.NotAfter,
                    MatchedKeyword: keyword.Value,
                    SourceLog:      s.sourceLog,
                })
            }
        }
    }
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
