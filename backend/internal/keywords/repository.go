package keywords

import (
    "context"
    "database/sql"
    "strings"
)

type Repository struct {
    db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
    return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context) ([]Keyword, error) {
    rows, err := r.db.QueryContext(ctx, `SELECT id, value, created_at FROM keywords ORDER BY value ASC`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    result := make([]Keyword, 0)
    for rows.Next() {
        var keyword Keyword
        if err := rows.Scan(&keyword.ID, &keyword.Value, &keyword.CreatedAt); err != nil {
            return nil, err
        }
        result = append(result, keyword)
    }

    return result, rows.Err()
}

func (r *Repository) Create(ctx context.Context, value string) (Keyword, error) {
    normalizedValue := strings.ToLower(strings.TrimSpace(value))

    var keyword Keyword
    err := r.db.QueryRowContext(
        ctx,
        `INSERT INTO keywords (value) VALUES ($1)
         ON CONFLICT (value) DO UPDATE SET value = EXCLUDED.value
         RETURNING id, value, created_at`,
        normalizedValue,
    ).Scan(&keyword.ID, &keyword.Value, &keyword.CreatedAt)

    return keyword, err
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
    _, err := r.db.ExecContext(ctx, `DELETE FROM keywords WHERE id = $1`, id)
    return err
}
