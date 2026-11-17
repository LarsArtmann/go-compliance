package compliance

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DefaultEvidenceFile is the fleet-standard piped-evidence location: the
// daemon recipe writes `buildflow compliance --json` there because the
// unprivileged overview service cannot read BuildFlow's state directory.
const DefaultEvidenceFile = "/var/lib/overview/compliance.json"

// EnvEvidenceFile overrides the Record default target.
const EnvEvidenceFile = "GO_COMPLIANCE_EVIDENCE_FILE"

// RunResult is what a reporter records for one repository run. Zero
// values are filled: At defaults to the record time, Runs to 1, FailRate
// to Failures/StepRows (or Failures/Runs when no step rows), and the nil
// maps/slices to their empty wire shapes.
type RunResult struct {
	// At is when the run finished.
	At time.Time
	// Verdict is the run's outcome (passed or failed).
	Verdict Verdict
	// Runs is the number of runs this result aggregates (default 1).
	Runs int64
	// StepRows is the number of step rows recorded.
	StepRows int64
	// Failures is the number of failed step rows.
	Failures int64
	// FailRate is failures over opportunities; derived when zero.
	FailRate float64
	// GateFindings is the per-step finding count ({step: count}).
	GateFindings map[string]int
	// FindingSeverities is the severity rollup ({severity: count}); the
	// v1 wire contract cannot carry it, so file-backed writers drop it
	// (the hub cache keeps it).
	FindingSeverities map[string]int
	// ProducerVersion identifies the reporting system.
	ProducerVersion string
}

// EvidenceWriter is the write-side port for reporters: somewhere a
// repository's run evidence can be recorded.
type EvidenceWriter interface {
	WriteEvidence(ctx context.Context, repo RepoName, result RunResult) error
}

// Record writes one repository's run evidence to the default fleet
// document: $GO_COMPLIANCE_EVIDENCE_FILE when set, else
// /var/lib/overview/compliance.json. It is the one-liner integration:
//
//	err := compliance.Record(ctx, compliance.NewRepoName("my-repo"),
//		compliance.RunResult{Verdict: compliance.VerdictPassed})
//
// The repository's row is replaced when the result is at least as new as
// the stored last run; older results are ignored (parallel reporters
// cannot regress the document). Severities are dropped: the frozen v1
// contract has no field for them.
func Record(ctx context.Context, repo RepoName, result RunResult) error {
	path := os.Getenv(EnvEvidenceFile)
	if path == "" {
		path = DefaultEvidenceFile
	}

	return RecordTo(ctx, NewFileStore(path), repo, result)
}

// RecordTo writes one repository's run evidence to an explicit writer.
func RecordTo(ctx context.Context, writer EvidenceWriter, repo RepoName, result RunResult) error {
	return writer.WriteEvidence(ctx, repo, result)
}

// MemoryStore is an in-process EvidenceWriter: the newest result per
// repository, kept as rich RepoEvidence. Safe for concurrent use.
type MemoryStore struct {
	mu       sync.Mutex
	evidence map[RepoName]RepoEvidence
	order    []RepoName
}

// NewMemoryStore returns an empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{evidence: make(map[RepoName]RepoEvidence)}
}

// WriteEvidence implements EvidenceWriter.
func (m *MemoryStore) WriteEvidence(_ context.Context, repo RepoName, result RunResult) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if existing, ok := m.evidence[repo]; ok && newerRun(existing.LastRunAt, result.resolvedAt()) {
		return nil
	}

	if _, ok := m.evidence[repo]; !ok {
		m.order = append(m.order, repo)
	}

	m.evidence[repo] = result.evidence(repo)

	return nil
}

// Evidence returns the stored evidence in first-write order.
func (m *MemoryStore) Evidence() []RepoEvidence {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make([]RepoEvidence, 0, len(m.order))
	for _, repo := range m.order {
		out = append(out, m.evidence[repo])
	}

	return out
}

// FileStore is a file-backed EvidenceWriter over the v1 fleet document:
// locked read-modify-write with an atomic rename. A missing file starts a
// fresh document.
type FileStore struct {
	path string
}

// NewFileStore returns a FileStore for the given document path.
func NewFileStore(path string) *FileStore { return &FileStore{path: path} }

// Path returns the document path.
func (s *FileStore) Path() string { return s.path }

// Load returns the current document; a missing file yields a fresh empty
// v1 document (a reporter may start one). Use ReadDocument when missing
// files must be distinguished.
func (s *FileStore) Load(ctx context.Context) (V1Report, error) {
	report, err := s.read(ctx)
	if err != nil {
		return V1Report{}, err
	}

	if report == nil {
		return emptyDocument(), nil
	}

	return *report, nil
}

// WriteEvidence implements EvidenceWriter.
func (s *FileStore) WriteEvidence(ctx context.Context, repo RepoName, result RunResult) error {
	lock, err := acquireFileLock(s.path + ".lock")
	if err != nil {
		return NewIOError(err, "lock evidence file %s", s.path)
	}
	defer func() { _ = lock.release() }()

	document, err := s.read(ctx)
	if err != nil {
		return err
	}

	if document == nil {
		fresh := emptyDocument()
		document = &fresh
	}

	if !mergeRecord(document, repo, result) {
		return nil
	}

	encoded, err := MarshalV1(*document)
	if err != nil {
		return err
	}

	if err := writeFileAtomic(s.path, encoded, 0o644); err != nil {
		return NewIOError(err, "write evidence file %s", s.path)
	}

	return nil
}

// read loads the document; a missing file yields (nil, nil).
func (s *FileStore) read(_ context.Context) (*V1Report, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil, nil
	}

	if err != nil {
		return nil, NewIOError(err, "read evidence file %s", s.path)
	}

	if len(data) == 0 {
		return nil, nil
	}

	report, err := DecodeV1Report(data)
	if err != nil {
		return nil, err
	}

	return &report, nil
}

// emptyDocument returns a schema-valid document with no repositories.
func emptyDocument() V1Report {
	return V1Report{
		SchemaVersion:    SchemaVersionV1,
		GeneratedAt:      FormatWireTime(time.Now()),
		StaleAfter:       DefaultStaleAfter.String(),
		Repos:            []V1Repo{},
		BuildFlowVersion: "go-compliance",
	}
}

// mergeRecord folds one run result into the document, newest-run-wins.
// Reports whether the document changed.
func mergeRecord(document *V1Report, repo RepoName, result RunResult) bool {
	at := result.resolvedAt()

	index := -1
	for i := range document.Repos {
		if document.Repos[i].Repo == repo.Get() {
			index = i
			break
		}
	}

	if index >= 0 && newerRun(ParseWireTime(document.Repos[index].LastRunAt), at) {
		return false
	}

	row := result.v1Repo(repo)

	if index < 0 {
		document.Repos = append(document.Repos, row)
	} else {
		document.Repos[index] = row
	}

	sortV1ReposWorstFirst(document.Repos)
	document.GeneratedAt = FormatWireTime(time.Now())

	return true
}

// resolvedAt returns the result time, defaulting to now.
func (r RunResult) resolvedAt() time.Time {
	if r.At.IsZero() {
		return time.Now()
	}

	return r.At
}

// newerRun reports whether stored is strictly newer than incoming.
func newerRun(stored, incoming time.Time) bool {
	return !stored.IsZero() && stored.After(incoming)
}

// evidence converts the result into rich RepoEvidence.
func (r RunResult) evidence(repo RepoName) RepoEvidence {
	runs, stepRows, failRate := r.resolvedCounters()

	return RepoEvidence{
		Repo:              repo,
		Runs:              runs,
		StepRows:          stepRows,
		Failures:          r.Failures,
		FailRate:          failRate,
		LastRunAt:         r.resolvedAt(),
		LastRunVerdict:    r.Verdict,
		GateFindings:      r.GateFindings,
		FindingSeverities: r.FindingSeverities,
		Versions:          versionsOf(r.ProducerVersion),
	}
}

// v1Repo converts the result into a wire row.
func (r RunResult) v1Repo(repo RepoName) V1Repo {
	runs, stepRows, failRate := r.resolvedCounters()

	return V1Repo{
		Repo:              repo.Get(),
		State:             DeriveState(r.evidence(repo), r.resolvedAt(), DefaultStaleAfter).String(),
		Runs:              runs,
		StepRows:          stepRows,
		Failures:          r.Failures,
		FailRate:          failRate,
		LastRunAt:         FormatWireTime(r.resolvedAt()),
		LastRunVerdict:    r.Verdict.String(),
		GateFindings:      orEmptyFindings(r.GateFindings),
		BuildFlowVersions: versionsOf(r.ProducerVersion),
		WorstSteps:        []V1Step{},
	}
}

// resolvedCounters fills the zero-valued counters.
func (r RunResult) resolvedCounters() (runs, stepRows int64, failRate float64) {
	runs, stepRows, failRate = r.Runs, r.StepRows, r.FailRate
	if runs == 0 {
		runs = 1
	}

	if failRate == 0 && r.Failures > 0 {
		if stepRows > 0 {
			failRate = float64(r.Failures) / float64(stepRows)
		} else {
			failRate = float64(r.Failures) / float64(runs)
		}
	}

	return runs, stepRows, failRate
}

func versionsOf(version string) []string {
	if version == "" {
		return []string{}
	}

	return []string{version}
}

func orEmptyFindings(findings map[string]int) map[string]int {
	if findings == nil {
		return map[string]int{}
	}

	return findings
}

// writeFileAtomic writes data to a temp file next to path and renames it
// into place, so readers never observe a torn document.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)

	temp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(temp.Name()) }()

	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return err
	}

	if err := temp.Close(); err != nil {
		return err
	}

	if err := os.Chmod(temp.Name(), perm); err != nil {
		return err
	}

	return os.Rename(temp.Name(), path)
}

// fileLock is an advisory exclusive lock on a sidecar file.
type fileLock struct{ file *os.File }

// acquireFileLock opens (creating) the lock file and takes an exclusive
// flock, blocking until held.
func acquireFileLock(path string) (*fileLock, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}

	if err := flockExclusive(file); err != nil {
		_ = file.Close()
		return nil, err
	}

	return &fileLock{file: file}, nil
}

// release drops the lock and closes the file.
func (l *fileLock) release() error {
	if err := flockRelease(l.file); err != nil {
		_ = l.file.Close()
		return err
	}

	return l.file.Close()
}
