# N400 Civics Study

An offline, single-binary Go study app built from the supplied official USCIS
PDFs. Phase 1 and the first six chapters' content milestones are implemented;
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
- `/learn`, `/learn/constitution`, `/learn/legislative`, `/learn/executive`,
  `/learn/judicial`, `/learn/rights`, and `/learn/geography`: Chapters 1–6's
  source-verified text editions, learning objectives, source-page markers,
  contents navigation, sidebar, diagram/map text, and 25, 20, 12, 8, 14, and 5
  bidirectional question links (Q18 links to two chapters, Q41 to three, Q13
  to three, Q50, Q64, and Q81 each to two).
- Five reviewed Study Guide images with their printed captions and credits.
  Chapters 4–6 have none: none of their photographs carry a reviewed printed
  credit (one of Chapter 5's five has a printed credit from a non-federal
  source, deferred pending a licensing review). Source coverage by page,
  parsed chapter goldens, reader HTML goldens, and manifest/file/reference
  checks run without PDFs or poppler.

Not implemented: Chapters 7–12, library transcription, remaining image curation, official/state
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
re-extracts only the five reviewed Chapter 1–3 images using an explicit allowlist.
They are retained at their source resolution (about 4.5 MB total). If a source PDF
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

Chapter 3 (The Executive Branch) was visually reviewed against PDF pages 24–28
the same way: prose, both repeated three-branch diagrams, the 22-item
Cabinet-level list, and the Line of Succession diagram all match the source,
and every diagram has an extractable PDF text layer (no visual supplement entry
needed, unlike Chapter 2). Of the seven images on these pages, only the 1964
voting photo (page 26) carries a printed credit (Library of Congress) and is
reproduced; the 2025 Cabinet photo, both presidential portraits, the FBI/USCIS
and armed-forces seals, and the party icons have no printed credit and are
omitted, with printed captions retained as reader text where any exist.
Omitting the sitting President's photo also sidesteps the same shelf-life
concern as the eight changing answers. The elephant icon's caption is printed
without a trailing period, unlike the donkey icon's; both are kept exactly as
printed.

Chapter 4 (The Judicial Branch) was visually reviewed against PDF pages 29–32
the same way: prose, the repeated three-branch diagram, the Federal Court
System pyramid, the Chief Justice duties list, and the majority sidebar all
match the source, with extractable PDF text throughout. None of its five
photographs (Thomas and Scalia, the Lady Justice statue, the courtroom, and
the nine-justices group photo) carry a printed credit, so none are reproduced;
printed captions are retained as reader text. This chapter's page 31 reuses
the chapter title verbatim as a genuine section heading, and a stray "CHAPTER
1: THE U.S. CONSTITUTION" running header bleeds onto its page 29 — both real
source artifacts, not transcription mistakes. They exposed a bug in the
source-coverage test's page-furniture filter (it was stripping the genuine
page-31 heading and missing the stray page-29 banner); the filter was
generalized rather than special-cased for this one chapter, and re-run against
Chapters 1–3 to confirm no change to their existing goldens.

Chapter 5 (Rights and Responsibilities) was visually reviewed against PDF pages
33–37 the same way, and the review caught a real transcription error before
commit: a first draft dropped "in federal elections" from one sentence, and
the source-coverage test failed immediately with the exact missing words —
concrete proof the test earns its keep. One sentence about the Federalist
Papers splits across pages 33/34 in the source itself; this edition reproduces
that split rather than smoothing it over. Of the chapter's five photographs,
four have no printed credit, and the fifth — a 2008 voting-booth photo — is
credited to "the Polling Place Photo Project," a non-federal source not yet
reviewed for redistribution terms; none of the five are reproduced, and that
one's full caption (credit line included) is kept as reader text pending a
deliberate licensing review, distinct from the other four's flat absence of
any credit. Two more source wording issues are kept exactly as printed:
"Civils Rights Act of 1964" (a typo for "Civil") and "will raise their right
hand say the Oath of Allegiance" (missing "and").

Chapter 6 (U.S. Geography) was visually reviewed against PDF pages 38–42 the
same way and turned up the most PDF-extraction quirks of any chapter so far —
all resolved by cross-checking the rendered page against the raw text, not
guessed. Three map labels follow a curved or diagonal path (page 39's
"Washington, D.C." marker, page 40's "GULF OF AMERICA," page 41's
mountain-range and river names) and extract as scrambled letter fragments
rather than legible words; each is handled through the same visual-supplement
mechanism as Chapter 2's page 22 diagram. Separately, page 41's PDF text layer
turned out to hide a full duplicate ALL-CAPS copy of "Alaska," "Hawaii," and
all five territory names behind the visible mixed-case labels — confirmed
against the rendered page to have only one printed instance of each — so the
coverage test drops one duplicate of each rather than this edition
transcribing labels that aren't actually there. This chapter's own running
header also breaks form, using an en dash ("CHAPTER 6 – U.S. GEOGRAPHY")
where Chapters 1–5 used a colon; the coverage test's banner filter was
broadened to accept either, re-confirmed against Chapters 1–5's existing
goldens. None of Chapter 6's five map graphics are reproduced as images:
none carry a printed credit, and full state/territory geography is not
reproduced, matching Chapter 1's map policy.

Next: author Chapter 7 and continue through the remaining chapters and their
question mappings, then the reference library. Full 128-question chapter coverage
remains a release gate; the first six chapters cover 25, 20, 12, 8, 14, and 5
questions (Q18 links to Chapters 1–2, Q41 to Chapters 1–3, Q13 to Chapters
1, 3, and 5, Q50 to Chapters 1 and 4, Q64 to Chapters 2 and 5, Q81 to
Chapters 1 and 6).

Initial verification: race tests, `go vet`, the native build, and all six
cross-platform builds passed. The binary was HTTP-smoke-tested from outside the
repository. The theme redesign was checked in Chromium: light/dark rendering,
saved theme across navigation, Auto tracking system appearance, and no horizontal
overflow at 380px on the home page, question list, and answer reader. Automatic
browser launch has not been manually verified. The Chapter 1 milestone also passed
race tests, vet, the native build, and Chromium checks at desktop and 380px widths
for contents anchors, all three images, question links, and chapter backlinks.
The Chapter 2–6 milestones passed the same race/vet/build checks and goldens;
their content was verified by direct visual PDF review above rather than a
separate Chromium pass.
The locally cached Staticcheck revision could not analyze Go 1.27's
export format; full `make lint` remains unverified pending a compatible tool.

Almanac text citation (required when its extracted text is reused): U.S.
Department of Homeland Security, U.S. Citizenship and Immigration Services,
Office of Citizenship, *The Citizen's Almanac*, Washington, DC, 2014.
