package compliance

import (
	"context"
	"os"
	"path/filepath"
	"time"
)

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

	if err := writeFileAtomic(s.path, encoded, evidenceFilePerm); err != nil {
		return NewIOError(err, "write evidence file %s", s.path)
	}

	return nil
}

// read loads the document; a missing file yields (nil, nil).
//
//nolint:nilnil // (nil, nil) is the documented missing/empty-file contract
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

const (
	// evidenceFilePerm: fleet evidence is world-readable data, not secrets.
	evidenceFilePerm = 0o644
	lockDirPerm      = 0o755
)

// writeFileAtomic writes data to a temp file next to path and renames it
// into place, so readers never observe a torn document.
//
//nolint:wrapcheck // callers wrap with NewIOError at the semantic boundary
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
//
//nolint:wrapcheck // callers wrap with NewIOError at the semantic boundary
func acquireFileLock(path string) (*fileLock, error) {
	if err := os.MkdirAll(filepath.Dir(path), lockDirPerm); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, evidenceFilePerm)
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
//
//nolint:wrapcheck // callers wrap with NewIOError at the semantic boundary
func (l *fileLock) release() error {
	if err := flockRelease(l.file); err != nil {
		_ = l.file.Close()

		return err
	}

	return l.file.Close()
}
