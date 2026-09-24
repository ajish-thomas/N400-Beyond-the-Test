# Project guidance: USCIS Civics Study Application

**Project:** A single-binary Go web app for studying the USCIS civics test
(N-400 naturalization), built entirely from official USCIS source materials.

**Design principle:** Fix correctness and accuracy bugs before adding scope.
Keep documentation honest when behavior changes. Ask when requirements conflict
or a product decision is needed.

See `CLAUDE.md` for project overview and settled technology decisions.
See `PLAN.md` for the full implementation specification.

## Implementation standards

### Go
- Idiomatic, readable Go. Prefer clarity over cleverness. Standard library first.
- Wrap errors with context: `fmt.Errorf("parsing question %d: %w", id, err)`.
- Focused functions, clear responsibilities, no mixing of concerns.
- `go vet` and staticcheck clean. Tests pass under `-race`.
- **No cgo.** It breaks cross-compilation to a single binary.
- **No Node, npm, or JS bundler.** `go build` is the entire build. Vendor JS
  libraries as static files under `internal/web/static/`.
- Keep the dependency list near-empty. Justify every third-party module.

### Content accuracy
- **Source of truth:** all study content derives from the official PDFs. Never
  invent, paraphrase, "improve", or supplement a fact from your own knowledge.
- **Transcribe, don't summarize.** Chapter prose is reproduced faithfully.
- **Validation:** verify against the original materials. Flag any divergence.
- **Completeness:** no information lost in translation from PDF to app.
- If a fact seems wrong in the source, surface it — do not silently correct it.

### Diagrams
- Use a D2 rendering only for a source diagram whose labels and relationships
  can be reproduced faithfully. Never add nodes, links, explanations, or
  inferred relationships; retain the complete source-text transcription in the
  reader as the accessible fallback and comparison record.
- Keep the D2 source in `internal/content/data/diagrams/` and check in both
  generated SVG variants under `internal/web/static/diagrams/`. Render them
  through the existing light/dark image pairing; do not rely on a single SVG to
  adapt itself at runtime.
- The SVG canvas must be transparent. The reader supplies the page background
  through `.d2-diagram { background: var(--bg) }`, so it exactly follows the
  selected theme.
- D2's bundled dark theme is purple and is not part of this product's visual
  system. Generate with `make diagrams` (using `D2=/path/to/d2` when needed):
  its required palette mapping produces app-consistent dark assets — charcoal
  surfaces (`#23231f`/`#2b2c26`), gray borders and labels (`#505147`/`#b4b2a3`),
  cream text (`#eeeadd`), and coral accents/arrows (`#f19c77`). Do not bypass
  that generation step or hand-edit a generated SVG.
- After adding or changing a diagram, update the applicable reader and content
  goldens, then run `make test`.

## Rules that must not be broken

These encode real constraints discovered during planning. Violating any of them
is a correctness or licensing bug, not a style preference.

**Answers are sets with a required count.**
- Each question has a list of *acceptable* answers and a `RequiredCount`.
- Most are 1 (any single listed answer is fully correct). Seven are not:
  `Q10:2, Q48:2, Q65:3, Q67:2, Q69:2, Q81:5, Q126:3`.
- A number in the question wording does **not** imply the count. Q16 "Name the
  three branches" is answered by one bullet; Q19 "two parts of Congress" by one
  bullet. Never derive the count by regex — it is authored in
  `required_counts.json` and guard-tested.
- For counts > 1, enforce **distinctness**: the same item named five times is
  one item, not five.

**Answer-string syntax.**
- `(parentheses)` mark **optional** text — `(U.S.) Constitution` must accept
  both "Constitution" and "U.S. Constitution".
- `[brackets]` are **reader instructions, never acceptable answers** — never
  grade user input against bracketed text.

**Grading is advisory, never punitive.**
- The real test is oral and judged by a USCIS officer. Normalize generously.
- **Every question carries an "I got this right" override.** The user is the
  authority. Never remove this escape hatch.

**The eight changing answers** (Q23, 29, 30, 38, 39, 57, 61, 62):
- The `uscis.gov/citizenship/testupdates` banner is shown **always** — whether
  the value is bundled, freshly fetched, or hand-entered. None of our sources is
  USCIS. Never blur that line.
- Manual override > sidecar file > live fetch > bundled snapshot. Always in that
  order.

**Offline-first.**
- Startup never blocks on network. No lookup on a hot path, ever.
- All network calls are user-initiated, short-timeout, and non-fatal on failure:
  keep the previous value and report the failure.
- `--offline` disables every network code path.

**Licensing and attribution.**
- **No images from `CitizensAlmanac-M-76.pdf`** — its notice states they are not
  public domain and may not be used outside that publication.
- **No AP-credited image** from the Study Guide.
- Every image displays its printed caption and credit line.
- Almanac *text* is reusable **with citation**: *U.S. Department of Homeland
  Security, U.S. Citizenship and Immigration Services, Office of Citizenship,
  The Citizen's Almanac, Washington, DC, 2014.*

**Privacy.**
- State, ZIP and progress stay on the local machine.
- A street address (optional, for district disambiguation) goes to the Census
  Bureau and is **never persisted**. Say so in the UI, plainly.

**ZIPs are not districts.** Some ZIPs span up to four congressional districts.
When one does, present the candidates and let the user choose. Never guess.

## Testing

**Tests must be hermetic.** No PDFs, no poppler, no live network, no clock
dependence. CI runs `go test -race ./...` and nothing else.

- **Content parsing** reads checked-in `data/raw/*.txt`, not the PDFs. Golden
  files pin the full parse so a parser change cannot silently corrupt content.
- **Network clients** are tested against `httptest` servers replaying recorded
  fixtures. Cover: happy path, 500, timeout, malformed JSON, empty result, and
  schema drift — each asserting the previous value survives and the failure is
  reported, not swallowed.
- **Time** is injected as an interface. Never call `time.Now()` in logic under test.
- **Randomness** is seeded. Quiz selection must be reproducible in tests.

Required coverage:

| Area | Must verify |
|---|---|
| Questions | 128 present, IDs 1–128 contiguous, exactly 20 asterisk-marked |
| **Grading** | **Every official answer, for all 128 questions, grades correct when submitted verbatim.** The most important test in the codebase. |
| Enumerations | N distinct passes; N−1 fails; N duplicates fail on distinctness |
| Optional parens | Both the core and the full form pass |
| Cardinality guard | Any prompt with a cardinality word is in the counts table or a reviewed-exceptions list |
| Cross-references | Every question→chapter, chapter→image, and library link resolves |
| Images | Manifest matches embedded files both ways; every entry credited; no AP |
| Officials | Overlay precedence; expiry boundary; multi-district ZIP; DC/territory wording |
| Handlers | Every route via `httptest`; 404s; golden HTML for reader and results |
| Store | Round-trip; atomic write; corrupt file recovers rather than crashes |

Also: manual testing of user workflows before release; document external
dependencies (poppler for ingest) and build steps.

## Workflow

- `make test` (`go test -race ./...`), `make lint`, `make build`, `make ingest`.
- `cmd/ingest` is **dev-only** and requires poppler (`pdftotext`, `pdfimages`).
  Its output is checked into the repo so nothing else depends on poppler.
- Content validation runs both as a test **and** at server startup.
- Never commit a change that makes the content tests fail. They are the contract
  with the source material.
