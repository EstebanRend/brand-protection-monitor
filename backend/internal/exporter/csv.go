package exporter

import (
    "encoding/csv"
    "io"
    "strconv"

    "brand-protection-monitor/backend/internal/certificates"
)

func WriteMatchesCSV(writer io.Writer, matches []certificates.MatchedCertificate) error {
    csvWriter := csv.NewWriter(writer)
    defer csvWriter.Flush()

    if err := csvWriter.Write([]string{"id", "domain", "issuer", "not_before", "not_after", "matched_keyword", "source_log", "created_at"}); err != nil {
        return err
    }

    for _, match := range matches {
        notBefore := ""
        if match.NotBefore != nil {
            notBefore = match.NotBefore.Format("2006-01-02T15:04:05Z07:00")
        }

        notAfter := ""
        if match.NotAfter != nil {
            notAfter = match.NotAfter.Format("2006-01-02T15:04:05Z07:00")
        }

        if err := csvWriter.Write([]string{
            strconv.FormatInt(match.ID, 10),
            match.Domain,
            match.Issuer,
            notBefore,
            notAfter,
            match.MatchedKeyword,
            match.SourceLog,
            match.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
        }); err != nil {
            return err
        }
    }

    return csvWriter.Error()
}
