package certificates

import "time"

type MatchedCertificate struct {
    ID             int64      `json:"id"`
    Domain         string     `json:"domain"`
    Issuer         string     `json:"issuer"`
    NotBefore      *time.Time `json:"notBefore"`
    NotAfter       *time.Time `json:"notAfter"`
    MatchedKeyword string     `json:"matchedKeyword"`
    SourceLog      string     `json:"sourceLog"`
    CreatedAt      time.Time  `json:"createdAt"`
}

type NewMatchedCertificate struct {
    Domain         string
    Issuer         string
    NotBefore      time.Time
    NotAfter       time.Time
    MatchedKeyword string
    SourceLog      string
}
