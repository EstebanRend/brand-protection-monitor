package monitor

import (
    "context"
    "database/sql"
)

type StateRepository struct {
    db *sql.DB
}

func NewStateRepository(db *sql.DB) *StateRepository {
    return &StateRepository{db: db}
}

func (r *StateRepository) Get(ctx context.Context) (State, error) {
    var state State
    err := r.db.QueryRowContext(ctx, `
        SELECT last_tree_size, last_processed_at, processed_last_cycle, status
        FROM monitor_state
        WHERE id = 1
    `).Scan(&state.LastTreeSize, &state.LastProcessedAt, &state.ProcessedLastCycle, &state.Status)
    return state, err
}

func (r *StateRepository) Update(ctx context.Context, treeSize int64, processed int, status string) error {
    _, err := r.db.ExecContext(ctx, `
        UPDATE monitor_state
        SET last_tree_size = $1,
            processed_last_cycle = $2,
            status = $3,
            last_processed_at = NOW(),
            updated_at = NOW()
        WHERE id = 1
    `, treeSize, processed, status)
    return err
}
