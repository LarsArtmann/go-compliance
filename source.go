package compliance

import (
	"context"
	"time"
)

// Source is a read-side evidence adapter: something that can produce
// per-repo compliance evidence for an evidence window. Implementations
// range from a piped v1 document (FileSource), a spawned producer CLI
// (CommandSource), to the hub's own SQLite cache.
type Source interface {
	// Name identifies the source for display and provenance.
	Name() SourceName
	// QueryEvidence returns per-repo evidence for the [since, now)
	// window. worstStepLimit caps per-repo worst-steps (<= 0 → default).
	// stepFilter scopes the evidence to ONE step by exact name where the
	// source supports it; empty means all steps. Sources that cannot
	// scope (a frozen v1 document is already aggregated) ignore it.
	QueryEvidence(
		ctx context.Context,
		since time.Time,
		worstStepLimit int,
		stepFilter string,
	) ([]RepoEvidence, error)
}

// SourceFunc adapts a function to the Source interface.
type SourceFunc func(ctx context.Context, since time.Time, worstStepLimit int, stepFilter string) ([]RepoEvidence, error)

// Name implements Source with a generic name.
func (SourceFunc) Name() SourceName { return NewSourceName("func") }

// QueryEvidence implements Source by calling the function.
func (f SourceFunc) QueryEvidence(
	ctx context.Context, since time.Time, worstStepLimit int, stepFilter string,
) ([]RepoEvidence, error) {
	return f(ctx, since, worstStepLimit, stepFilter)
}
