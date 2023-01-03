-- name: AddLambdaStat :exec
INSERT INTO lambda_stat(project, name, started_at, finished_at, environment, error)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING id;

-- name: ListLambdaStats :many
SELECT *
FROM lambda_stat
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: ListProjectLambdaStats :many
SELECT *
FROM lambda_stat
WHERE project = ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: ListProjectSpecificLambdaStats :many
SELECT *
FROM lambda_stat
WHERE project = ?
  AND name = ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: GCLambdaStats :exec
DELETE
FROM lambda_stat
WHERE started_at < ?;