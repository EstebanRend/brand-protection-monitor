CREATE TABLE IF NOT EXISTS keywords (
    id BIGSERIAL PRIMARY KEY,
    value TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS matched_certificates (
    id BIGSERIAL PRIMARY KEY,
    domain TEXT NOT NULL,
    issuer TEXT NOT NULL,
    not_before TIMESTAMPTZ,
    not_after TIMESTAMPTZ,
    matched_keyword TEXT NOT NULL,
    source_log TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(domain, matched_keyword, source_log)
);

CREATE TABLE IF NOT EXISTS monitor_state (
    id BIGINT PRIMARY KEY DEFAULT 1,
    last_tree_size BIGINT NOT NULL DEFAULT 0,
    last_processed_at TIMESTAMPTZ,
    processed_last_cycle INT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'idle',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT only_one_monitor_state CHECK (id = 1)
);

INSERT INTO monitor_state (id, status)
VALUES (1, 'idle')
ON CONFLICT (id) DO NOTHING;
