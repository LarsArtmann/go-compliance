package compliance

import id "github.com/larsartmann/go-branded-id"

// RepoBrand brands repository names so they cannot be mixed with other
// string-shaped identifiers at compile time.
type RepoBrand struct{}

// Name implements the branded-id naming interface ("Repo:go-finding").
func (RepoBrand) Name() string { return "Repo" }

// RepoName identifies one repository in the fleet (e.g. "go-finding").
// The zero value means "unset" and serializes to JSON null.
type RepoName = id.ID[RepoBrand, string]

// NewRepoName brands a raw repository name.
func NewRepoName(raw string) RepoName { return id.NewID[RepoBrand](raw) }

// SourceBrand brands evidence-source names.
type SourceBrand struct{}

// Name implements the branded-id naming interface ("Source:buildflow").
func (SourceBrand) Name() string { return "Source" }

// SourceName identifies one evidence source (e.g. "buildflow", "hub-cache").
// The zero value means "unset" and serializes to JSON null.
type SourceName = id.ID[SourceBrand, string]

// NewSourceName brands a raw source name.
func NewSourceName(raw string) SourceName { return id.NewID[SourceBrand](raw) }
