// Command redactfixture reads a raw `buildflow compliance --json` document
// on stdin and writes a redacted, canonical v1 fixture to stdout: every
// repository name is replaced with a stable placeholder (repo-0001, ...,
// in first-appearance order), including inside step fan-out keys like
// "samber-linter [work/<repo>]". All other fields pass through unchanged
// (real timestamps, counts, states, ordering). The output is our canonical
// marshal, so the committed fixture doubles as the deterministic-marshal
// golden.
//
// Not part of the public API; refresh via scripts/refresh-fixture.sh.
package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/larsartmann/go-compliance"
)

func main() {
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fatal("read stdin", err)
	}

	report, err := compliance.DecodeV1Report(raw)
	if err != nil {
		fatal("decode fixture", err)
	}

	aliases := map[string]string{}

	for i := range report.Repos {
		name := report.Repos[i].Repo
		if name == "" {
			continue
		}

		if _, exists := aliases[name]; !exists {
			aliases[name] = fmt.Sprintf("repo-%04d", len(aliases)+1)
		}

		report.Repos[i].Repo = aliases[name]
	}

	// Longest first: "go-health" must not clobber "go-health-dashboard".
	names := make([]string, 0, len(aliases))
	for name := range aliases {
		names = append(names, name)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(names)))
	sort.SliceStable(names, func(i, j int) bool { return len(names[i]) > len(names[j]) })

	redactString := func(s string) string {
		for _, name := range names {
			s = strings.ReplaceAll(s, name, aliases[name])
		}
		return s
	}

	for i := range report.Repos {
		repo := &report.Repos[i]

		redactedFindings := make(map[string]int, len(repo.GateFindings))
		for step, count := range repo.GateFindings {
			redactedFindings[redactString(step)] = count
		}
		repo.GateFindings = redactedFindings

		for j := range repo.WorstSteps {
			repo.WorstSteps[j].Step = redactString(repo.WorstSteps[j].Step)
		}
	}

	encoded, err := compliance.MarshalV1(report)
	if err != nil {
		fatal("encode redacted fixture", err)
	}

	if _, err := os.Stdout.Write(encoded); err != nil {
		fatal("write stdout", err)
	}

	fmt.Fprintln(os.Stderr, "redacted", len(aliases), "repository name(s)")
}

func fatal(what string, err error) {
	fmt.Fprintf(os.Stderr, "redactfixture: %s: %v\n", what, err)
	os.Exit(1)
}
