# N400 Civics Study

An offline, single-binary Go study app built from the supplied official USCIS
PDFs. Phase 1 and the first two chapters' content milestones are implemented;
this is not a release of the full plan.

## Run

Requires Go 1.22 or newer. No third-party Go modules, Node, bundler, cgo, runtime
PDFs, or network connection are required.

```sh
make build
./bin/n400
# Or keep the browser closed:
./bin/n400 --offline --no-browser --port 8400
```

Open http://127.0.0.1:8400. The server binds only to loopback, validates all
embedded questions, chapters, image credits, and cross-references at startup,
and shuts down on Ctrl+C. You can move the binary
to another directory and run it without this repository. `go build ./cmd/n400`
also produces the complete app. `make build-all` builds Linux, macOS, and Windows
for amd64 and arm64 with cgo disabled.

## Implemented

- All 128 official questions, full answer menus, source notes, and 20 starred
  questions; browse all or the 65/20 set and reveal answers.
- Authored required counts, separate optional text and reader guidance, startup
  validation, and a full-parse golden comparison against `data/questions.json`.
- Permanent USCIS verification links on all eight changing questions. These
  currently show the PDF's instructions, **not current officeholder names**.
- A field-guide-inspired interface with paper light and charcoal dark themes,
  a Light / Dark / Auto selector saved locally in the browser, keyboard-accessible answer reveals and an
  “I got this right” self-check. Self-checks are not persisted yet.
- Development extraction of all four PDFs to checked-in raw text.
- `/learn`, `/learn/constitution`, and `/learn/legislative`: Chapters 1–2's
  source-verified text editions, learning objectives, source-page markers,
  contents navigation, sidebar, diagram/map text, and 25 plus 20 bidirectional
  question links (Q18 links to both).
- Four reviewed Study Guide images with their printed captions and credits.
  Source coverage by page, parsed chapter goldens, reader HTML goldens, and
  manifest/file/reference checks run without PDFs or poppler.

Not implemented: Chapters 3–12, library transcription, remaining image curation, official/state
snapshots and lookups, onboarding, automated grading, practice tests, Leitner
scheduling, and progress storage. No network client exists; `--offline` is
accepted now and must guard future clients. No personal information is collected.
HTMX/Alpine will be vendored when interactive workflows need them; this initial
reader uses native HTML with a small, embedded script for theme selection and
no third-party JavaScript dependency. Auto follows the system appearance;
without JavaScript the interface also follows the system appearance.

## Develop and verify

```sh
make test    # go test -race ./...; no PDFs, poppler, network, or clock required
make lint    # go vet and staticcheck (install staticcheck separately)
make build
make ingest  # dev only: requires poppler's pdftotext
```

`cmd/ingest` expects the four PDFs named in `PLAN.md` at the repository root.
`--source` overrides their directory. It extracts `internal/content/data/raw/`
and generates `internal/content/data/questions.json`. Review any generated diff
against the original source before accepting a new golden; do not blindly update
it to fix a failing test. Tests consume checked-in text, JSON, Markdown, and images.
Text is reflowed across PDF lines; punctuation and wording are retained.

For local image curation, `go run ./cmd/ingest --images` also requires `pdfimages`.
It extracts **only the Study Guide** into ignored `data/images/_extracted/`.
This unreviewed directory may contain AP material: do not publish or embed it.
Almanac images are never extracted. `go run ./cmd/ingest --curated-images`
re-extracts only the four reviewed Chapter 1–2 images using an explicit allowlist.
They are retained at their source resolution (about 3.1 MB total). If a source PDF
changes, recheck the image identity, printed caption, credit, and licensing before
accepting new output; extraction indices alone are not provenance.

Chapter authoring uses JSON front matter (a YAML subset) and a small Markdown
subset, rendered with escaped Go templates and no Markdown dependency. See
`internal/content/data/chapters/README.txt`. To update goldens after source review:

```sh
go test ./internal/content -run TestChapterGolden -update-chapter-golden
go test ./internal/web -run TestChapterReaderGoldenAndImages -update-reader-golden
```

## Source review notes and next steps

The supplied Q&A PDF is M-1778 (09/25). Compared with the planning notes, its Q48
has **22** answer bullets (not 19), and Q67 has **six** (not five). Their required
counts remain two. The source includes “Secretary of War (Defense)”; it is
preserved exactly, not silently corrected. The actual section boundaries are
Q16–62 for System of Government and Q63–72 for Rights and Responsibilities.
Q117's standalone tribal-list note is retained separately from its final answer.
Q120's bracketed “Also acceptable…” passage remains reader guidance per the
project rule, and is displayed intact.

Before Phase 3 grading, resolve a specification conflict: an individual official
bullet must be recognized, but a single bullet cannot pass a question requiring
N distinct items. The proposed test contract is that every bullet self-matches
as an item, while whole-question success still requires N distinct items.
“Answers will vary” and “Visit…” are source instructions, not factual answers;
changing questions need resolved answers before automated grading.

Chapter 1 was visually reviewed against PDF pages 8–17. All prose, objectives,
sidebar text, diagram/map labels, and captions are retained. Diagrams are text
transcriptions; the map's geography is not reproduced. Uncredited illustrations
are omitted and their printed captions remain visible. Page 17 is blank except
its footer. This is explicitly labeled a text edition, not a complete facsimile.
Source wording, including “Respresentatives” and “is a called a governor,” remains
unchanged and is flagged in the reader's source notes. The chapter-coverage check
compares word inventories per source page; visual review and the golden separately
pin reading order, which word counts alone cannot prove.

Chapter 2 was visually reviewed against PDF pages 18–23 the same way: prose, both
“U.S. Congress” tree diagrams, the lawmaking flowchart, the credited Oval Office
photo and its credit, and blank page 23 all match the source. The uncredited
Capitol photo on page 20 is omitted; its printed caption is retained as reader
text. The line-break hyphen in “representa-tives” on page 18 is joined. Page 22's
lawmaking diagram has no extractable PDF text layer; its labels were transcribed
separately into `data/raw/study-guide-visual-supplement.json` and are checked by
the same source-coverage test as the rest of the chapter.

Next: author Chapter 3 and continue through the remaining chapters and their
question mappings, then the reference library. Full 128-question chapter coverage
remains a release gate; the first two chapters cover 25 and 20 questions
(Q18 is shared between them).

Initial verification: race tests, `go vet`, the native build, and all six
cross-platform builds passed. The binary was HTTP-smoke-tested from outside the
repository. The theme redesign was checked in Chromium: light/dark rendering,
saved theme across navigation, Auto tracking system appearance, and no horizontal
overflow at 380px on the home page, question list, and answer reader. Automatic
browser launch has not been manually verified. The Chapter 1 milestone also passed
race tests, vet, the native build, and Chromium checks at desktop and 380px widths
for contents anchors, all three images, question links, and chapter backlinks.
The Chapter 2 milestone passed the same race/vet/build checks and goldens; its
content was verified by direct visual PDF review above rather than a separate
Chromium pass.
The locally cached Staticcheck revision could not analyze Go 1.27's
export format; full `make lint` remains unverified pending a compatible tool.

Almanac text citation (required when its extracted text is reused): U.S.
Department of Homeland Security, U.S. Citizenship and Immigration Services,
Office of Citizenship, *The Citizen's Almanac*, Washington, DC, 2014.
