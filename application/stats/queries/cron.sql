-- name: AddCronStat :exec
INSERT INTO cron_stat(project, expression, started_at, finished_at, error)
VALUES (?, ?, ?, ?, ?)
RETURNING id;

-- name: ListCronStats :many
SELECT *
FROM cron_stat
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: ListProjectCronStats :many
SELECT *
FROM cron_stat
WHERE project = ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: ListProjectSpecificCronStats :many
SELECT *
FROM cron_stat
WHERE project = ? AND expression = ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: GCCronStats :exec
DELETE
FROM cron_stat
WHERE started_at < ?;