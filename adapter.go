package compliance

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

// Producer defaults mirroring the overview consumer's exec contract: a
// bounded read (a misbehaving producer is an error, not an OOM) and a
// source-scoped timeout.
const (
	DefaultCommandTimeout = 30 * time.Second
	DefaultCommandMaxOut  = 64 << 20 // 64 MiB
)

// FileSource reads a piped v1 document from disk - the daemon-mode
// recipe: a system timer writes `buildflow compliance --json` to
// /var/lib/overview/compliance.json and every consumer reads the file.
type FileSource struct {
	name SourceName
	path string

	// missingTTL backs off repeated reads of a missing file: after a
	// miss, reads inside the TTL return the cached miss without touching
	// disk. Zero (default) retries every read.
	missingTTL time.Duration

	mu         sync.Mutex
	lastMiss   time.Time
	cachedMiss error
}

// NewFileSource returns a FileSource for the given path.
func NewFileSource(path string) *FileSource {
	return &FileSource{name: NewSourceName("file"), path: path}
}

// NewBuildFlowFileSource returns a FileSource for the fleet-standard
// piped BuildFlow evidence document.
func NewBuildFlowFileSource() *FileSource {
	source := NewFileSource(DefaultEvidenceFile)
	source.name = NewSourceName("buildflow")

	return source
}

// WithMissingBackoff sets how long a missing file is remembered and returns
// the source for chaining. Inside the TTL after a miss, reads return the
// cached miss without touching disk, so a refresher that queries every few
// seconds does not poll the disk for a producer that is not running. A ttl
// of zero or less (the default) caches nothing: every read consults the
// file, so the first evidence shows with no delay. Set it before the source
// serves queries; the TTL must not change while reads are in flight.
func (s *FileSource) WithMissingBackoff(ttl time.Duration) *FileSource {
	s.missingTTL = ttl

	return s
}

// Name implements Source.
func (s *FileSource) Name() SourceName { return s.name }

// Path returns the document path.
func (s *FileSource) Path() string { return s.path }

// QueryEvidence implements Source. The frozen document is already
// aggregated, so stepFilter is ignored (documented Source contract) and
// worstStepLimit can only truncate the delivered worst-steps lists.
func (s *FileSource) QueryEvidence(
	_ context.Context, _ time.Time, worstStepLimit int, _ string,
) ([]RepoEvidence, error) {
	if missed := s.missBackoff(); missed != nil {
		return nil, missed
	}

	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		miss := NewIOError(err, "evidence file %s does not exist yet (is the producer timer running?)", s.path)
		s.rememberMiss(miss)

		return nil, miss
	}

	if err != nil {
		return nil, NewIOError(err, "read evidence file %s", s.path)
	}

	report, err := DecodeV1Report(data)
	if err != nil {
		return nil, err
	}

	return V1ReportToEvidence(report, worstStepLimit), nil
}

// missBackoff returns the cached miss when inside the TTL.
func (s *FileSource) missBackoff() error {
	if s.missingTTL <= 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cachedMiss != nil && time.Since(s.lastMiss) < s.missingTTL {
		return s.cachedMiss
	}

	return nil
}

// rememberMiss caches a missing-file miss for the backoff window.
func (s *FileSource) rememberMiss(miss error) {
	if s.missingTTL <= 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastMiss = time.Now()
	s.cachedMiss = miss
}

// V1ReportToEvidence normalizes a decoded wire document into rich
// RepoEvidence. Fields the contract cannot carry (severities, repo hash,
// last-run producer version) are absent; states re-derive identically
// from the normalized evidence when using the document's own
// GeneratedAt/StaleAfter (golden-tested against a real producer).
func V1ReportToEvidence(report V1Report, worstStepLimit int) []RepoEvidence {
	worstStepLimit = worstStepLimitOrDefault(worstStepLimit)

	evidence := make([]RepoEvidence, 0, len(report.Repos))

	for _, row := range report.Repos {
		record := RepoEvidence{
			Repo:           NewRepoName(row.Repo),
			Runs:           row.Runs,
			StepRows:       row.StepRows,
			Failures:       row.Failures,
			FailRate:       row.FailRate,
			LastRunAt:      ParseWireTime(row.LastRunAt),
			LastRunVerdict: Verdict(row.LastRunVerdict),
			GateFindings:   row.GateFindings,
			Versions:       row.BuildFlowVersions,
		}

		steps := make([]StepEvidence, 0, len(row.WorstSteps))
		for _, step := range row.WorstSteps {
			steps = append(steps, StepEvidence(step))
		}

		if len(steps) > worstStepLimit {
			steps = steps[:worstStepLimit]
		}

		record.WorstSteps = steps
		evidence = append(evidence, record)
	}

	return evidence
}

// CommandSource spawns a producer CLI and decodes its v1 JSON from
// stdout - the integration mode for producers that own their evidence
// store (BuildFlow: `buildflow compliance --json`).
type CommandSource struct {
	name      SourceName
	command   string
	args      []string
	timeout   time.Duration
	maxOutput int64
}

// NewCommandSource returns a CommandSource running command with args
// (which must select the v1 JSON output, e.g. "compliance", "--json").
func NewCommandSource(command string, args ...string) *CommandSource {
	return &CommandSource{
		name:      NewSourceName(baseName(command)),
		command:   command,
		args:      args,
		timeout:   DefaultCommandTimeout,
		maxOutput: DefaultCommandMaxOut,
	}
}

// NewBuildFlowCommandSource returns a CommandSource spawning the
// BuildFlow CLI with the v1 contract flags.
func NewBuildFlowCommandSource() *CommandSource {
	return NewCommandSource("buildflow", "compliance", "--json")
}

// WithTimeout overrides the exec timeout.
func (s *CommandSource) WithTimeout(timeout time.Duration) *CommandSource {
	s.timeout = timeout

	return s
}

// WithMaxOutput overrides the bounded-read limit.
func (s *CommandSource) WithMaxOutput(limit int64) *CommandSource {
	s.maxOutput = limit

	return s
}

// Name implements Source.
func (s *CommandSource) Name() SourceName { return s.name }

// QueryEvidence implements Source, mapping the window onto the producer's
// flags: --since (the window length), --worst-steps, and --step for the
// exact-name step filter.
func (s *CommandSource) QueryEvidence(
	ctx context.Context, since time.Time, worstStepLimit int, stepFilter string,
) ([]RepoEvidence, error) {
	if s.timeout > 0 {
		var cancel context.CancelFunc

		ctx, cancel = context.WithTimeout(ctx, s.timeout)
		defer cancel()
	}

	// Headroom for the query flags appended below: --since, --worst-steps,
	// and --step each contribute a name and a value.
	const maxQueryFlagArgs = 6

	args := make([]string, len(s.args), len(s.args)+maxQueryFlagArgs)
	copy(args, s.args)

	if !since.IsZero() {
		args = append(args, "--since", time.Since(since).Round(time.Second).String())
	}

	if worstStepLimit > 0 {
		args = append(args, "--worst-steps", strconv.Itoa(worstStepLimit))
	}

	if stepFilter != "" {
		args = append(args, "--step", stepFilter)
	}

	//nolint:gosec // the command and base flags are operator configuration
	cmd := exec.CommandContext(ctx, s.command, args...)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, NewIOError(err, "%s %v failed (stderr: %s)", s.command, s.args, truncateForLog(stderr.String()))
	}

	output, err := io.ReadAll(io.LimitReader(&stdout, s.maxOutput))
	if err != nil {
		return nil, NewIOError(err, "read %s output", s.command)
	}

	report, err := DecodeV1Report(output)
	if err != nil {
		return nil, err
	}

	return V1ReportToEvidence(report, worstStepLimit), nil
}

// truncateForLog bounds stderr context in error messages.
func truncateForLog(text string) string {
	const limit = 200
	if len(text) <= limit {
		return text
	}

	return text[:limit] + "..."
}

// baseName returns the final path element of a command path.
func baseName(command string) string {
	for i := len(command) - 1; i >= 0; i-- {
		if command[i] == '/' {
			return command[i+1:]
		}
	}

	return command
}
