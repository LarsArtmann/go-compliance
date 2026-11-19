# go-compliance — Status Report (2026-11-17 20:44 CET)

Scope: execution session against the Pareto plan
(`docs/planning/2026-11-17_20-01_go-compliance-pareto-execution-plan.md`),
starting at M1/F001. Method: one medium task at a time, build+test green
before moving on, honest reporting of everything that broke along the way.

---

## a) FULLY DONE

### Tier 1 — core interchange library (the 1% → 51%)

- **M1 (F001–F008) state model**: `state.go` (State/Verdict types, 4 states
  incl. the new `never-covered`, `DeriveState` port of BuildFlow's
  `DeriveComplianceState` with the zero-evidence → never-covered extension),
  `branded.go` (RepoName/SourceName via go-branded-id v0.7.1 phantom types),
  `errors.go` (5 sentinels + `*compliance.Error` implementing
  ErrorCode/ErrorFamily — the go-finding FindingError pattern), `evidence.go`
  (`RepoEvidence`/`StepEvidence`/`HasEvidence`), ginkgo suite bootstrap,
  48-spec table-driven suite green.
- **M2 (F009–F012) v1 wire types**: `v1.go` — `V1Report`/`V1Repo`/`V1Step`
  byte-faithful to BuildFlow's `complianceReportJSON` (field order, snake_case
  keys, nil-map `{}` / nil-slice `[]` normalization), `MarshalV1` with
  `json.Deterministic(true)`, `DecodeV1Report` (schema_version gate),
  `BuildV1Report` (port of `buildComplianceReport` incl. worst-first sort),
  `FormatWireTime`/`ParseWireTime` helpers. Round-trip + exact-shape tests.
- **M3 (F013–F017) golden compat, both directions**:
  - `scripts/redactfixture/main.go` + `scripts/refresh-fixture.sh`: capture
    live `buildflow compliance --json` (341 repos, 177 KB) and redact all
    repository names (first-appearance `repo-0001…`, including inside step
    fan-out keys, longest-first replacement). Verified clean of known names
    (eventcatalog-hub, BuildFlow, overview, desire-secrets, invoices, …).
  - `testdata/buildflow-v1-fixture.json` committed (redacted, canonical
    marshal).
  - Golden tests: marshal(decode(fixture)) == fixture bytes exactly; only the
    three v1 producer states occur (never-covered never appears);
    overview-mirror rejects `schema_version: 2`; truncated documents rejected.
- **Killer compat proof (discovered, not planned)**:
  `V1ReportToEvidence` + `DeriveState` **re-derive all 341 producer states
  identically** when run at the document's own GeneratedAt/StaleAfter — the
  state model is proven equivalent on real data, not just unit tests.

### Tier 4 — usable CLI surface (part of the 4% → 64%)

- **M4 (F018–F021) evidence fold**: `fold.go` — `FoldEvidence` port of
  BuildFlow's ensure/union fold (window health overwrite, latest-run
  overwrite, worst-steps cap w/ default 3, findings/severities attach-only
  non-empty, **versions attach-only never mint**), name-sorted deterministic
  output; quiet repos never dropped.
- **M5 (F022–F025) Source + reporter SDK**: `source.go` (Source interface +
  SourceFunc), `reporter.go` (`RunResult`, `EvidenceWriter`, `Record`/
  `RecordTo` one-liner, `MemoryStore`, `FileStore` with flock + atomic
  temp-rename + newest-run-wins merge), godoc `ExampleRecord`. Race-detector
  green incl. concurrent writers/readers.
- **M6 (F026–F030) adapters**: `adapter.go` — `FileSource` (piped v1 doc,
  missing-file backoff TTL), `NewBuildFlowFileSource`
  (/var/lib/overview/compliance.json preset), `CommandSource` (CLI spawn,
  bounded read 64 MiB, 30s timeout, stderr-context errors,
  window→`--since/--worst-steps/--step` flag mapping),
  `NewBuildFlowCommandSource`.
- **M7 (F031–F035) hub module + cmdguard CLI**: `go.work` (core + hub),
  `hub/go.mod` (cmdguard v5.1.0, dev-time `replace ../` documented for
  release-time removal), `cmd/compliance-hub/main.go` (taskctl pattern:
  `v5.ExitCode(run(...))`), `internal/cli/`: app.go (NewCLI + report/table/
  serve), flags.go (QueryFlags/ServeFlags with flag-parity tags +
  `parseSinceDuration`/`parseStaleAfter`/`parseDurationSupportingDays`
  ports), report.go (v1 JSON producer), table.go, serve.go (minimal HTTP:
  /healthz + /v1/compliance, graceful ctx shutdown). 6-spec CLI smoke suite
  green (report decodes, table renders, flag validation errors name allowed
  values, missing-file guidance, serve+healthz over a real socket).
- **M8 (F036–F039) go-output fleet table**: `colors.go` (lipgloss color
  service, NO_COLOR + colorprofile auto, one authority for cells+chrome),
  table.go rewritten on go-output/table: REPO/STATE/ERR/WARN/INFO/FINDINGS/
  RUNS/FAILS/RATE/VERDICT/LAST RUN columns, severity cells dash-for-empty,
  state labels colored, worst-first via BuildV1Report sort, 24-rune repo
  truncation with ellipsis. **Gotcha-#298 regression test**: ANSI-colored
  rows must share identical box-separator positions with the header —
  caught a real test-design bug (byte vs rune offsets) and now pins display
  alignment.

### Tier 20 (part of the 20% → 80%)

- **M9 (F040–F044) sqlite evidence cache**: `hub/internal/cache/cache.go` —
  go-sqlitestore v0.1.1 Open + membership migrations (version 1:
  `evidence_runs` append-only + repo/time index), `WriteEvidence`
  (EvidenceWriter impl), `QueryEvidence` (Source impl folding window health,
  all-history latest runs, latest-run findings/severities JSON columns,
  window-distinct versions). 5-spec suite green: round-trip, newest-verdict,
  quiet-repo stale derivation, Source/EvidenceWriter interface conformance,
  reopen/migration idempotence.
- **M10 core half (F047)**: `CompleteWithDiscovery` in core `fold.go` —
  unions evidence with discovered repo names, minting zero-evidence rows
  that derive as never-covered; deterministic name-sorted output.

### Research (subagents, all reports consumed)

- go-health-dashboard composition (go-sse/go-datastar/templ-components/
  RegisterRoutes/go-health/go-daemon APIs with exact signatures).
- cmdguard v5 usage (NewCLI/NewCommand/AddCommands/ExecuteWithArgs, flag
  tags, taskctl composition) — verified against source before use.
- go-sqlitestore/go-branded-id/go-error-family APIs + project-discovery-
  daemon wire contract (socket, /v1/discover, /v1/watch SSE, domain.Event
  JSON).

---

## b) PARTIALLY DONE

- **M10 (F045–F049) discovery client — code written, NOT compiling** (see d):
  `hub/internal/discovery/client.go` + `client_test.go` exist; the client
  (unix-socket HTTP, Discover, Watch SSE parser, RepoNames) and a full
  unix-socket test server are written, but the package fails `go vet`/build
  on two defects listed in (d). ~15 min of fixes + test-run from green.
- **M7 serve**: functional but minimal — serves a startup snapshot as JSON;
  the realtime internals (M11–M15, M27) are designed to replace it. Flag
  surface (`--addr`, localhost default) is final per the bind posture.
- **CLI source selection**: `--source file|buildflow` only; the cache (M9)
  is not yet reachable from the CLI (`--source cache --db …` lands with
  M14's content negotiation work).

---

## c) NOT STARTED

- M11 go-sse broadcaster + /events (F050–F053)
- M12 go-datastar patch stream (F054–F057)
- M13 templ-components fleet table UI (F058–F062)
- M14 filter chips + RegisterRoutes embed + content negotiation (F063–F067)
- M15 go-health probe + go-daemon sd_notify + systemd unit (F068–F071)
- M27 watch-mode live re-derive (F107–F109)
- M20 bind-posture decision doc + unix-socket listen mode (F090–F093; the
  localhost default + `--addr` flag already exist from M7)
- M17 docs set: dprint.json, TODO_LIST.md, FEATURES.md, DOMAIN_LANGUAGE.md,
  CHANGELOG.md (F077–F081)
- M18 flake.nix via go-nix-helpers (F082–F085)
- M19 CI per module + drift check (F086–F089)
- M21 integration guide + examples/reporter + screenshots (F094–F096)
- M22 benchmarks at 300 repos (F097–F098)
- M23 BuildFlow TODO_LIST pointer + STACK errata (F099–F100)
- M24 v2 API sketch (F101–F102)
- M25 federation research (F103–F104)
- M26 overview endgame note (F105–F106; blocked on owner answer g2)
- Core doc.go mentions `CommandSource`/`SourceFunc` — both exist; but
  `ApplySimpleFixes`-style README docs for the core API do not exist yet
  (README is still the pre-inception stub).

---

## d) TOTALLY FUCKED UP (all caught, none shipped green)

1. **discovery package does not compile** (`go vet` error):
   `client.go:120: undefined: json.NewDecoder` — I called `json.NewDecoder`
   on the `encoding/json/v2` package, whose reader-based decoding API is
   different (v2 uses `jsontext` for stream decoding). Fix: read the body
   (`io.ReadAll`) then `json.Unmarshal`, matching the pattern used
   everywhere else in this repo.
2. **discovery test file has a nonsense helper**: `client_test.go` ends with
   `func jsonNewDecoder(r io.Reader) *jsonDecoder { return newJSONDecoder(r) }`
   referencing types that do not exist (leftover from an abandoned approach);
   also missing `io`/`encoding/json/v2` imports. Fix: delete the helper and
   decode the request body directly in the handler.
3. **Fixed during the session (kept honest)**: ginkgo DescribeTable needs
   `time.Duration(0)` not untyped `0` (reflection type mismatch); golden
   byte-compare assumed a trailing newline that the fixture does not have;
   `sh -c` swallows the first following arg as `$0` (flag-mapping test);
   `os.IsNotExist` fires for missing parent dirs too (IO-error test now uses
   an ENOTDL path); alignment test compared BYTE offsets although the
   ellipsis cell is multi-byte (now compares rune positions — this was a
   test bug, the table was display-correct all along); vet printf-check
   rejected non-constant format strings in `NewValidationError` calls
   (switched to `%s` args).
4. **Edit-tool races with the auto-commit daemon** (recurring): the daemon's
   format pass repeatedly touched files between my read and edit, voiding
   edits ("file modified since last read"). Worked around via python
   heredoc patches three times. No content was lost; the workaround is
   recorded here because it will bite again in M13's generated templ files.

---

## e) WHAT WE SHOULD IMPROVE (in this repo, next sessions)

1. **`queryLatestRunCounts` is N+1** — one query per repo for latest-run
   findings/severities. At 341 repos that is ~700 queries per report.
   Fine for a single-shot CLI, wrong for the hub's watch mode; fold into
   one query with a window join before M27.
2. **Cache lacks per-step rows** — WorstSteps are structurally empty from
   the cache Source and stepFilter cannot scope. If fleet value materializes,
   add an `evidence_steps` table (migration v2) instead of overloading
   gate_findings.
3. **Severity cells unreachable via file source** — the v1 wire cannot carry
   severities (frozen contract), so the CLI table renders dashes for piped
   documents; only cache/reporter paths can fill them. The UI (M13) should
   read from the cache, not from v1 files, for the same reason.
4. **Two sources of truth for "producer version"**: hub `cli.Version = "dev"`
   vs the report's `buildflow_version` field. Set via ldflags at release;
   add a version-check script before tagging (go-finding precedent).
5. **Serve captures evidence once at startup** — a long-running serve serves
   a stale snapshot until M27 lands; document this in the M11 rebuild.
6. **Edit-tool vs daemon races**: prefer `write` (full file) or python
   patches for generated/large files; check `git status` before authoring
   commit messages (daemon absorbed staged files mid-session repeatedly).
7. **Fixture pin**: `v1_golden_test.go` asserts exactly 341 repos — the
   refresh script will change that count as the fleet moves; relax to
   `> 300` when refreshing or pin per-fixture.
8. **No lint config yet** — golangci/dprint configs land with M17/M19; until
   then `gofmt -w .` is the only formatting gate.

---

## f) NEXT UP TO 50 TASKS (ranked; parent IDs from the plan)

1. Fix discovery compile: io.ReadAll + json.Unmarshal in Discover (d1)
2. ~~Delete bogus jsonNewDecoder helper in client_test.go (d2)~~ done — helper deleted from hub/internal/discovery/client_test.go
3. Run discovery suite green (unix-socket httptest server is already written)
4. M11/F050: broadcaster wiring + heartbeat (go-sse NewBroadcaster + Stream)
5. M11/F051: evidence refresh → state-diff events
6. M11/F052: /events endpoint with Last-Event-ID replay (EventStore ring)
7. M11/F053: endpoint tests (ssetest or hand-rolled SSE client)
8. M12/F054: patch types for fleet table rows (go-datastar ElementsPatch)
9. ~~M12/F055: signal patches for counts/badges (SignalsPatch)~~ done — dashboard sendSignals SignalsPatch
10. ~~M12/F056: ScriptHandler serving DataStar JS~~ done — dashboard ScriptHandler serves the pinned bundle
11. ~~M12/F057: datastartest E2E assertions~~ done — datastartest assertions live in dashboard + cli E2E tests
12. M13/F058: templ setup + layout via templ-components (go:generate, tool
    directive, commit generated files)
13. M13/F059: fleet table component (repo/state/severity columns)
14. M13/F060: badges + stat cards
15. M13/F061: detail row expansion (worst steps)
16. M13/F062: page render test
17. M14/F063: filter chips (state filter)
18. ~~M14/F064: grouping by severity/state~~ done — state-ranked sort hub/internal/cli/table.go + ?state= filter hub/internal/dashboard/routes.go
19. M14/F065: RegisterRoutes(mux) public surface
20. M14/F066: content negotiation HTML/JSON
21. ~~M14/F067: embedding quickstart in README~~ done — README links docs/guides/integrate.md
22. M14: `--source cache --db path` CLI wiring for M9
23. M15/F068: go-health probe registration
24. M15/F069: sd_notify wiring via go-daemon
25. M15/F070: systemd unit file + docs
    ~~26. M15/F071: smoke: serve + curl /health~~ done at 2026-11-19: serve smoke exists — dynamic port, /healthz + full route sweep with hardening-header assertions (hub/internal/cli/app_test.go:96-150)
26. M27/F107: re-derive on store writes (hook)
27. M27/F108: debounce + coalesce events
28. M27/F109: E2E: evidence change → browser event
29. M27: discovery /v1/watch → refresh trigger (client.Watch exists after d1–d3)
30. M20/F090: decision doc docs/decisions/bind-posture.md
31. M20/F091: bind flag + localhost default (flag exists; document+test)
32. M20/F092: optional unix-socket listen mode
33. M20/F093: test non-localhost bind requires explicit flag
34. M17/F077: dprint.json + format check
35. M17/F078: TODO_LIST.md seeded from the plan (HARVEST target)
36. ~~M17/F079: FEATURES.md initial inventory~~ done — FEATURES.md
37. ~~M17/F080: DOMAIN_LANGUAGE.md core terms~~ done — docs/DOMAIN_LANGUAGE.md
38. M17/F081: CHANGELOG.md Unreleased section
39. M18/F082–F085: flake.nix skeleton + devShell + packages + `nix flake
    check` green
40. M19/F086–F089: CI build/test/lint per module, drift-check, docs gates,
    Forgejo+GitHub runner validity
41. M21/F094–F096: examples/reporter, one-line integration guide, screenshots
42. M22/F097–F098: benchmarks (fold+derive at 300 repos) + baseline file
43. M23/F099–F100: BuildFlow TODO_LIST pointer row + STACK.md errata check
44. M24/F101–F102: v2 API design doc + compat policy note
45. M25/F103–F104: go-health-dashboard federation research + interop note
46. M26/F105–F106: overview endgame note + migration sketch (blocked: g2)
47. ~~Fix queryLatestRunCounts N+1 before M27 (e1)~~ done — queryLatestRecords single correlated-subquery query hub/internal/cache/cache.go
48. Relax the 341-repo fixture pin before the next fixture refresh (e7)
49. Final: whole-repo build+test+lint+docs audit vs plan, then release
    procedure (drop hub's replace directive, tag core v0.1.0 first)

---

## g) UP TO 3 QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Forgejo visibility (carried over)**: is `lars/go-compliance` set public
   in the Forgejo UI? Push-to-create used the server default and there is no
   API token/tea CLI to check or flip it (anonymous API 404s even for known
   repos). Needs one manual click by the owner.
2. **g1 — hub auth/bind posture**: localhost-only is the safe default (and
   is what ships). Should the hub ever bind the LAN (plain HTTP? unix socket
   - reverse proxy? auth token?), and if so which mechanism is the fleet
     standard? Blocks M20's unix-socket mode value and any remote dashboard
     use.
3. **g2 — overview endgame**: replace overview's compliance views with the
   hub (embed via RegisterRoutes + migration sketch), or keep overview as
   the consumer and the hub as a parallel producer forever? Blocks M26;
   until answered both dashboards will drift features independently.

---

Machine state at report time: core module `go test ./...` GREEN (107+ specs);
hub cli+cache suites GREEN; hub discovery package RED on compile (d1/d2);
working tree has uncommitted `fold.go` (CompleteWithDiscovery) + untracked
`hub/internal/discovery/` — the auto-commit daemon will absorb them.

**WAITING for the user. No further work in progress.**
