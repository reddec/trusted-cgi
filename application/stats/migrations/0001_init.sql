-- +migrate Up
CREATE TABLE endpoint_stat
(
    id               BIGINT PRIMARY KEY AUTOINCREMENT NOT NULL PRIMARY KEY,
    project          TEXT                             NOT NULL,
    method           TEXT                             NOT NULL,
    path             TEXT                             NOT NULL DEFAULT '',
    request_url      TEXT                             NOT NULL DEFAULT '',
    started_at       TIMESTAMP                        NOT NULL DEFAULT current_timestamp,
    finished_at      TIMESTAMP                        NOT NULL DEFAULT current_timestamp,
    request_headers  TEXT                             NOT NULL, -- JSON
    request_body     BLOB                             NOT NULL, -- Could be truncated: len(body) < size
    request_size     BIGINT                           NOT NULL,
    response_headers TEXT                             NOT NULL, -- JSON
    response_body    BLOB                             NOT NULL, -- Could be truncated: len(body) < size
    response_size    BIGINT                           NOT NULL,
    status           INTEGER                          NOT NULL
);

CREATE INDEX endpoint_stat_project ON endpoint_stat (project);
CREATE INDEX endpoint_stat_project_method_path ON endpoint_stat (project, method, path);

CREATE TABLE cron_stat
(
    id          INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL PRIMARY KEY,
    project     TEXT                              NOT NULL,
    expression  TEXT                              NOT NULL,
    started_at  TIMESTAMP                         NOT NULL DEFAULT current_timestamp,
    finished_at TIMESTAMP                         NOT NULL DEFAULT current_timestamp,
    error       TEXT                              NOT NULL DEFAULT ''
);
CREATE INDEX cron_stat_project ON cron_stat (project);
CREATE INDEX cron_stat_project_expression ON cron_stat (project, expression);

CREATE TABLE lambda_stat
(
    id          INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL PRIMARY KEY,
    project     TEXT                              NOT NULL,
    name        TEXT                              NOT NULL,
    started_at  TIMESTAMP                         NOT NULL DEFAULT current_timestamp,
    finished_at TIMESTAMP                         NOT NULL DEFAULT current_timestamp,
    environment TEXT                              NOT NULL, -- JSON
    error       TEXT                              NOT NULL DEFAULT ''
);

CREATE INDEX lambda_stat_project ON lambda_stat (project);
CREATE INDEX lambda_stat_project_expression ON lambda_stat (project, name);

-- +migrate Down
DROP INDEX lambda_stat_project_expression;
DROP INDEX lambda_stat_project;
DROP TABLE lambda_stat;

DROP INDEX cron_stat_project_expression;
DROP INDEX cron_stat_project;
DROP TABLE cron_stat;

DROP INDEX endpoint_stat_project_method_path;
DROP INDEX endpoint_stat_project;
DROP TABLE endpoint_stat;