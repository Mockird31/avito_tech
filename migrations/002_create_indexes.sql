-- Таблица "user"
CREATE INDEX IF NOT EXISTS idx_user_team_name ON "user"(team_name);
CREATE INDEX IF NOT EXISTS idx_user_team_active_true ON "user"(team_name, id) WHERE is_active = TRUE;

-- Таблица pull_request_reviewers
CREATE INDEX IF NOT EXISTS idx_prr_reviewer_pr ON pull_request_reviewers(reviewer_id, pull_request_id);

-- Таблица pull_request
CREATE INDEX IF NOT EXISTS idx_pull_request_author ON pull_request(author_id);
CREATE INDEX IF NOT EXISTS idx_pull_request_name ON pull_request(name);