package stats

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
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
