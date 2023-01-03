-- name: AddEndpointStat :exec
INSERT INTO endpoint_stat
(project, method, path, request_url, started_at, finished_at, payload_size, headers, status, body, truncated)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id;

-- name: ListEndpointStats :many
SELECT *
FROM endpoint_stat
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: ListProjectEndpointStats :many
SELECT *
FROM endpoint_stat
WHERE project = ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: ListProjectSpecificEndpointStats :many
SELECT *
FROM endpoint_stat
WHERE project = ?
  AND method = ?
  AND path = ?
ORDER BY id DESC
LIMIT ? OFFSET ?;


-- name: GetEndpointStat :one
SELECT *
FROM endpoint_stat
WHERE id = ?;

-- name: GCEndpointStats :exec
DELETE
FROM endpoint_stat
WHERE started_at < ?;