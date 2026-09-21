# N400 Civics Study

An offline, single-binary Go study app built from the supplied official USCIS
PDFs. Phase 1 and all 12 chapters' content milestones are implemented; this
is not a release of the full plan.

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
  `/learn/judicial`, `/learn/rights`, `/learn/geography`, `/learn/early-history`,
  `/learn/revolution`, `/learn/new-government`, `/learn/civil-war`,
  `/learn/modern-history`, and `/learn/symbols-holidays`:
  all 12 chapters' source-verified text editions, learning objectives,
  source-page markers, contents navigation, sidebar, diagram/map text, and
  25, 20, 12, 8, 14, 5, 4, 11, 8, 7, 15, and 11 bidirectional question links
  (Q81 links to four chapters, Q41, Q13, Q86 each to three or four, Q14,
  Q18, Q50, Q58, Q64, Q82, Q91, Q122, and Q128 each to two, Q126 to three).
  107 of 128 official questions are linked to a chapter that literally
  states an accepted answer; the other 21 aren't stated verbatim by any
  chapter and are deliberately left unlinked rather than forced.
- Thirty-one reviewed Study Guide images with their printed captions and
  credits, including a National Park Service photo in Chapter 7, a U.S.
  Senate photo in Chapter 8, NASA and White House photos in Chapter 11, and
  a second U.S. Senate photo in Chapter 12 — new credit-allowlist additions
  since Chapter 2's JFK Library entry. Chapters 4–6 have none from the
  Study Guide itself: none of their photographs carry a reviewed printed
  credit (one of Chapter 5's five, and one of Chapter 7's four, have
  printed credits from non-federal sources, both deferred pending a
  licensing review; one of Chapter 8's seven, five of Chapter 9's six, four
  of Chapter 10's eight, five of Chapter 11's fifteen, and eleven of
  Chapter 12's fifteen, have no credit at all; one of Chapter 12's fifteen
  is credited to the Associated Press, the first case this project's
  standing wire-service rejection rule has actually excluded). Source
  coverage by page, parsed chapter goldens, reader HTML goldens, and
  manifest/file/reference checks run without PDFs or poppler.
- Five additional images — two verified public-domain reference maps (a
  National Archives 1775 colonies map, a USGS national reference map), the
  colonies map reused across Chapters 1, 6, 7, and 8 as four manifest
  entries — illustrate genuine geography the Study Guide's own uncredited
  maps can't supply as images. Sourced and license-verified individually,
  not from a "fair use" image search; see
  `internal/content/data/images/EXTERNAL-SOURCES.txt`.
- `/library`, `/library/declaration`, and `/library/us-constitution`: the
  reference library's first two documents, sharing their Markdown/JSON
  authoring format with Learn chapters (`internal/content/library.go` factors
  the shared parser out of `chapters.go`). The Declaration of Independence is
  transcribed in full from `DOI-Constitution-M-654.pdf` pages 7–13, including
  its signature block and its 13-state signers list, and links 5 questions
  (Q8, Q9, Q10, Q11, Q79). Q78 ("Who wrote the Declaration of Independence?")
  is deliberately not linked: the document names Thomas Jefferson only as a
  Virginia signer and never states that he drafted it. The Constitution's
  Preamble and Articles I–VII are transcribed from pages 15–38, including its
  own signers list (12 states — Rhode Island sent no delegates) and, for
  completeness, the ratifying convention's closing resolution and the first
  Congress's resolution transmitting the Bill of Rights, both of which the
  source prints immediately after the Articles and before the Amendments.
  Its 11 numbered footnotes (each marking text later superseded by a specific
  amendment) are inlined as a parenthetical at their point of reference
  rather than built as a separate footnote UI. It links 10 questions
  (Q2, Q5, Q16, Q17, Q19, Q20, Q25, Q36, Q42, Q50); several topically close
  questions (Cabinet, veto, "for life") are deliberately not linked because
  the 1787 text never uses those words. Source coverage by page and a parsed
  golden run the same way as chapters, against `data/raw/constitution.txt`.

Not implemented: the rest of the reference library (the Constitution's 27
Amendments, and the Citizen's Almanac's speeches, symbols and anthems, and
landmark Supreme Court cases), remaining image curation, official/state
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
re-extracts only the thirty-one reviewed Chapter 1–3 and 7–12 images from
the PDFs using an explicit allowlist. They are retained at their source
resolution (about 16 MB total). If a source PDF changes, recheck the image
identity, printed caption, credit, and licensing before accepting new
output; extraction indices alone are not provenance. The five additional
externally sourced maps (Chapters 1, 6, 7, 8) are outside this pipeline
entirely — they are not extracted from any supplied PDF — and are
re-verified against `internal/content/data/images/EXTERNAL-SOURCES.txt`, not
regenerated by `--curated-images`.

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

Chapter 7 (Early American History) was visually reviewed against PDF pages
43–46 the same way: prose, both extractable maps (the transatlantic-routes
map and the 13-colonies map), and the sidebar on the 1707 union of England
and Scotland all match the source. Two of its four photographs are
reproduced: a Jamestown street scene credited to the National Park Service —
a federal agency, added to the credit allowlist, the first addition since
Chapter 2's JFK Library entry — and a painting already covered by the
existing Library of Congress credit. The other two are omitted: one credited
to the Jamestown Yorktown Foundation, a Virginia state institution (deferred
pending review, same as Chapter 5's Polling Place Photo Project credit), and
one with no printed credit at all. Page 45's "Atlantic Ocean" map label
follows the ocean's curve and extracts as scrambled fragments, handled via
the same visual-supplement mechanism as several of Chapter 6's labels. This
chapter covers Native American depopulation and the start of slavery in the
colonies factually, in the source's own words, without added commentary.

Following that, Chapters 1, 6, and 7 were retrofitted with two verified
public-domain reference maps to illustrate geography their own Study Guide
maps can't supply as images (no printed credit). Each was found via search,
then independently confirmed public domain on its actual source page before
download — explicitly not a "fair use" internet image grab. The National
Archives' 1775 colonies map is reused at three chapter pages (one file, three
manifest entries, since a manifest image is pinned to exactly one page); the
USGS national reference map covers Chapter 6's borders and rivers content but
not its territories. Two other categories (a territories-with-capitals map;
a historical map of colonization and slave-trade Atlantic routes) had no
verifiable public-domain candidate — Library of Congress item pages, the most
promising source for both, returned HTTP 403 to automated fetching — and
remain text. Full source citations and license statements are recorded in
`internal/content/data/images/EXTERNAL-SOURCES.txt`.

Chapter 8 (The American Revolutionary War & The Declaration of Independence)
was visually reviewed against PDF pages 47–51 the same way, and is the
richest image chapter yet: all six of its Study Guide photographs/paintings
are reproduced, none omitted for lack of a credit. Three credits were
already allowlisted (National Archives, National Park Service, Library of
Congress); "Courtesy of the U.S. Senate." is a new addition, reviewed as a
federal legislative body's own art collection. Only the uncredited Benjamin
Franklin portrait is omitted. The Currier & Ives print's credit carries a
National Archives item number ("...National Archives, 532915."), moved into
the caption so the credit field holds the exact allowlisted string, losing
no text. A sentence spans the page 47/48 boundary in the source itself,
handled the same way as Chapter 5's Federalist Papers split. This chapter's
printed title splash is also the first to span two lines rather than one;
the coverage test's title-splash filter was generalized to match the full
span's concatenation, re-confirmed against Chapters 1–7's existing goldens.
The National Archives' 1775 Mitchell Map is reused here too (a fourth
manifest copy) alongside this chapter's own 13-states content.

Chapter 9 (A New Government and an Expanding Nation) was visually reviewed
against PDF pages 52–57 the same way. Only one of its six illustrations
carries a printed credit (the Constitution document, already allowlisted);
the other five (the Oath of Office painting, the Federalist title page, and
three uncredited maps) have none and are omitted, captions retained as
text. Unlike every image chapter since Chapter 1, none of its three map
graphics have any extractable text layer beyond their captions — no state
names or other labels — so no visual-supplement entry was needed here. This
chapter substantially repeats content from Chapters 1, 5, 6, and 8 (the
Constitutional Convention, the Federalist Papers, George Washington's
"Father of Our Country" epithet) with new specifics (exact presidential
dates); genuinely repeated facts are linked to every chapter that states
them. Its American Indian tribes list includes "Arawak" and "Inuit," both
absent from the official answer pool for that question — kept exactly as
printed, not trimmed to match.

Chapter 10 (The Civil War) was visually reviewed against PDF pages 58–61 the
same way. Four of its eight illustrations carry a printed credit and are
reproduced (cotton picking and Susan B. Anthony from the Library of Congress,
Frederick Douglass from the National Park Service, the Thomas Nast
"Emancipation" print from the Library of Congress); the other four (the
Battle of Antietam painting, the 1861 free/slave states map, and both
Lincoln images) have none and are omitted. The "Emancipation" print's own
credit line is missing the trailing period this chapter's other two Library
of Congress credits use; the manifest's credit field holds the standard
normalized string, and the punctuation difference is noted rather than
silently carried through. As with Chapter 9, none of this chapter's
uncredited illustrations have any extractable text layer beyond their
captions. Frederick Douglass has no dedicated question in the official 128
and is covered anyway, per this project's second goal of teaching the
material, not only the test answers.

Chapter 11 (American History: 1900-2001) was visually reviewed against PDF
pages 62–68 the same way. Ten of its fifteen illustrations carry a printed
credit and are reproduced; the other five (the Cold War map, Dr. Martin
Luther King Jr., Dolores Huerta, the March on Washington, and the September
11 rescue-workers photo) have none and are omitted. Two credit sources are
new to the allowlist: "Courtesy of NASA." (the Apollo 11 Moon-landing photo)
and "Courtesy of the White House." (the Pentagon flag photo). The Pentagon
flag photo's printed credit reads "White House photo by Paul Morse," with no
"Courtesy of..." phrasing in the source at all; normalizing it to the
allowlisted string introduces three words not literally printed, recorded in
the visual supplement for page 67. The Cold War map's "United States" and
"Soviet Union" labels are cleanly extractable and transcribed as a diagram;
its "Atlantic Ocean" label follows the ocean's curve and extracts as
scrambled fragments, handled the same way as several earlier chapters' map
labels. This chapter restates two facts already covered elsewhere with new
specifics that justify new links: the federal government's exclusive power
to declare war (Chapter 1 states it generally; this chapter narrates
Congress exercising it in 1941, dual-linking Q58) and the 22nd Amendment's
two-term limit (naming the amendment directly, unlike Chapter 3's mention of
the rule without naming its source).

Chapter 12 (American Symbols and Holidays) was visually reviewed against PDF
pages 69–76 the same way, completing all 12 chapters. Three of its fifteen
captioned illustrations carry a printed credit and are reproduced (the
Statue of Liberty and Abraham Lincoln from the Library of Congress,
"George Washington at Princeton" from the U.S. Senate); one more carries a
printed credit that is explicitly excluded rather than reproduced:
"Raising the Flag on Iwo Jima" is credited to "the Associated Press," which
this project's credit allowlist has rejected since Chapter 1 as a
commercial wire-service credit, not a federal institution's own collection
— this is the first chapter where that standing rule actually excludes an
image. The other eleven captioned illustrations have no printed credit at
all and are omitted the same way. Two further graphics (the chapter-opening
banner and the full flag illustration) have no caption at all and aren't
counted as illustrations, the same treatment given every chapter's
uncaptioned opening banner since Chapter 1. This chapter restates several
facts already covered elsewhere with new specifics: George Washington's
"Father of Our Country" epithet (Q86, now four chapters), the War of 1812
and the Civil War as 1800s wars (Q91), the 50 stars (Q122), Veterans Day
(Q128), and national holidays generally (Q126, now three chapters), for
which this chapter's own "U.S. Holidays" list is the definitive source. Q125
("What is Independence Day?") is a new link, stated directly by this
chapter's own Independence Day section; Q8 is deliberately not linked here
despite similar phrasing, for consistency with Chapter 8's existing scope.

Full 128-question chapter coverage was never expected to follow automatically
from finishing all 12 chapters and remains open: 107 of 128 questions are
linked to a chapter whose own prose states an accepted answer; the other 21
questions aren't stated verbatim anywhere in the 12 chapters and are
deliberately left unlinked. All 12 chapters cover 25, 20, 12, 8, 14, 5, 4,
11, 8, 7, 15, and 11 questions respectively (Q81 links to Chapters 1, 6, 7,
and 8; Q41, Q13, and Q86 each to three or four; Q14, Q18, Q50, Q58, Q64,
Q82, Q91, Q122, and Q128 each to two; Q126 to three).

The reference library is underway. `internal/content/chapters.go` was split
to expose a shared Markdown/JSON-front-matter parser (`splitDocument`,
`parseBlocks`), so `internal/content/library.go` reuses it for library
documents instead of a second parser; library documents carry `kind` and
`source` metadata instead of a chapter number, objectives, and images, and
an Almanac-sourced document must carry the exact required citation string
or it fails to parse. The Declaration of Independence is the first document,
transcribed from `DOI-Constitution-M-654.pdf` pages 7–13 and linking 5
questions (Q8–Q11, Q79); Q78 is deliberately not linked since the document
never states who wrote it. Its text is single-column, unlike the Study
Guide's two-column chapters, but three sentences still split across a page
boundary in the source and are authored as matching split blocks.

The Constitution's Preamble and Articles I–VII are the second document
(`us-constitution.md`, pages 15–38), also single-column. It includes the
Constitution's own signers list (12 states; Rhode Island sent no delegates
to the Convention) and, for completeness, two procedural pages the source
prints immediately after the Articles and before the Amendments: the
ratifying convention's closing resolution and the first Congress's
resolution transmitting the Bill of Rights. Eleven sentences split across a
page boundary in the source (more than the Declaration's three, since these
pages run longer and denser); each is authored as a matching split block.
The source prints 11 numbered footnotes marking text later superseded by a
specific amendment (e.g. the original method of electing Senators, changed
by the Seventeenth Amendment); each is inlined as a parenthetical at its
point of reference — "for six Years;" printed with a bare "2" marking
footnote 2 becomes "for six Years;" followed by "(Changed by the Seventeenth
Amendment.)" — rather than a separate footnote/superscript UI. Because the
reference digit itself would otherwise still count as a word in the source's
own text, `TestLibrarySourceCoverage` strips each of the 11 known reference
digits from the raw side via a small, exact-match map, the same targeted
technique chapters' `mapLabelFixes` uses for PDF-extraction artifacts. This
document links 10 questions (Q2, Q5, Q16, Q17, Q19, Q20, Q25, Q36, Q42,
Q50); several topically close questions (Cabinet, veto, "for life") are
deliberately not linked because the 1787 text never uses those words, and
none is linked for a fact whose current authority is a superseded, bracketed
clause pending the Amendments document.

Next: the Constitution's 27 Amendments, then the Citizen's Almanac's
speeches, symbols/anthems, and landmark cases.

Initial verification: race tests, `go vet`, the native build, and all six
cross-platform builds passed. The binary was HTTP-smoke-tested from outside the
repository. The theme redesign was checked in Chromium: light/dark rendering,
saved theme across navigation, Auto tracking system appearance, and no horizontal
overflow at 380px on the home page, question list, and answer reader. Automatic
browser launch has not been manually verified. The Chapter 1 milestone also passed
race tests, vet, the native build, and Chromium checks at desktop and 380px widths
for contents anchors, all three images, question links, and chapter backlinks.
The Chapter 2–12 milestones passed the same race/vet/build checks and goldens;
their content was verified by direct visual PDF review above rather than a
separate Chromium pass.
The locally cached Staticcheck revision could not analyze Go 1.27's
export format; full `make lint` remains unverified pending a compatible tool.

Almanac text citation (required when its extracted text is reused): U.S.
Department of Homeland Security, U.S. Citizenship and Immigration Services,
Office of Citizenship, *The Citizen's Almanac*, Washington, DC, 2014.
