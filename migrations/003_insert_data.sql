COPY team (name, created_at, updated_at)
FROM '/home/csv/teams.csv'
WITH (FORMAT csv, HEADER);

COPY "user" (id, username, team_name, is_active, created_at, updated_at)
FROM '/home/csv/users.csv'
WITH (FORMAT csv, HEADER);

COPY pull_request (id, name, author_id, status, created_at, merged_at)
FROM '/home/csv/pull_requests.csv'
WITH (FORMAT csv, HEADER, NULL 'NULL');

COPY pull_request_reviewers (pull_request_id, reviewer_id, created_at, updated_at)
FROM '/home/csv/pull_request_reviewers.csv'
WITH (FORMAT csv, HEADER, NULL 'NULL');