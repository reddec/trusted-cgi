package stats

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"github.com/hashicorp/go-multierror"
	migrate "github.com/rubenv/sql-migrate"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"time"
)

func RunGC(ctx context.Context, db DBTX, keep time.Duration) error {
	q := New(db)
	oldest := time.Now().Add(-keep)
	var errs *multierror.Error
	if err := q.GCEndpointStats(ctx, oldest); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("GC endpoint stats: %w", err))
	}
	if err := q.GCLambdaStats(ctx, oldest); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("GC lambda stats: %w", err))
	}
	if err := q.GCCronStats(ctx, oldest); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("GC cron stats: %w", err))
	}
	return errs.ErrorOrNil()
}

//go:embed migrations
var Migrations embed.FS

func Open(file string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", file)
	if err != nil {
		return nil, err
	}
	_, err = migrate.Exec(db, "sqlite3", migrate.EmbedFileSystemMigrationSource{
		FileSystem: Migrations,
		Root:       "migrations",
	}, migrate.Up)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}
