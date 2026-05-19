package certificates

import (
    "context"
    "database/sql"
)

type Repository struct {
    db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
    return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context) ([]MatchedCertificate, error) {
    rows, err := r.db.QueryContext(ctx, `
        SELECT id, domain, issuer, not_before, not_after, matched_keyword, source_log, created_at
        FROM matched_certificates
        ORDER BY created_at DESC
        LIMIT 500
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    result := make([]MatchedCertificate, 0)
    for rows.Next() {
        var item MatchedCertificate
        if err := rows.Scan(&item.ID, &item.Domain, &item.Issuer, &item.NotBefore, &item.NotAfter, &item.MatchedKeyword, &item.SourceLog, &item.CreatedAt); err != nil {
            return nil, err
        }
        result = append(result, item)
    }

    return result, rows.Err()
}

func (r *Repository) SaveMatch(ctx context.Context, item NewMatchedCertificate) error {
    _, err := r.db.ExecContext(ctx, `
        INSERT INTO matched_certificates (domain, issuer, not_before, not_after, matched_keyword, source_log)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (domain, matched_keyword, source_log) DO NOTHING
    `, item.Domain, item.Issuer, item.NotBefore, item.NotAfter, item.MatchedKeyword, item.SourceLog)

    return err
}
