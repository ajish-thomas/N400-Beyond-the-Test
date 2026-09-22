# N400 Civics Study App — Implementation Plan

## Implementation status — September 20, 2026

This section records the actual implementation. The phase descriptions below
remain the target specification; they are not claims that every feature exists.

| Area | Status | Implemented / remaining |
|---|---|---|
| Phase 1: skeleton and extraction | Implemented | Loopback Go server; embedded assets; graceful shutdown; `--port`, `--no-browser`, `--offline`; four raw PDF text extractions; validated 128-question parser; optional-text/guidance separation; full-parse golden. |
| Phase 2: chapters | All 12 chapters authored | Chapters 1–12 text editions are implemented and source-verified, with 25, 20, 12, 8, 14, 5, 4, 11, 8, 7, 15, and 11 question links respectively and thirty-six credited images total (five sourced externally as verified public-domain maps; Chapters 4–6 otherwise have none from the PDF itself). 107 of 128 official questions are linked to a chapter; the remaining 21 are not literally stated by any chapter's own prose and are deliberately unlinked rather than forced. The reference library (Phase 2's other component) remains. |
| Phase 2: required counts | Implemented | All seven enumeration counts are authored and guard-tested; actual grading is not implemented. |
| Phase 2: library and officials | In progress | The library authoring format and infrastructure are implemented, sharing chapters' Markdown/JSON parser via an extracted common core. Four source-verified documents are published: the Declaration of Independence, the Constitution's Preamble and Articles I-VII, Amendments I-XXVII, and — the first from the Citizen's Almanac — "Patriotic Anthems of the United States" (the national anthem, "America the Beautiful," and "The New Colossus"). The Almanac's remaining symbols (Pledge, Flag, Motto, Great Seal), seven speeches, and four landmark cases are deliberately deferred. A dated snapshot covers four federal offices, state capitals, both senators in every state, and all 435 House members plus D.C.'s delegate; the 119th Congress Census ZIP-to-district crosswalk is bundled; state QIDs enable a live, per-state governor refresh (not bundled — see below). |
| Phase 3: engines | In progress | Advisory grading, deterministic official/65-20 quiz-session selection, a five-box Leitner scheduler, and atomic local progress storage are implemented and race-tested. The federal and governor refresh clients, local sidecar persistence, the Settings refresh/reset actions, and the `--offline` guard on refresh are implemented and fixture/route-tested. The bundled House roster resolves Q29 from state + district (or from state alone for at-large/single-seat cases), and a Census address geocoder lets a learner resolve an exact district by street address instead of picking from a multi-district ZIP's candidate list. All eight changing questions now resolve to a concrete, source-labeled answer wherever their source data allows it. |
| Phase 4: web UI | Partially implemented | Home, question list/detail, 65/20 filter, answer reveal, self-check, Learn index, chapter reader, curated image routes, Library index/document reader, durable Flashcards, and official/65-20 Practice Test flows with local answer history are implemented. Welcome and Settings remain. |
| Phase 5: build and release | Partially implemented | Make targets and README exist; six-platform build checked at the initial milestone. CI configuration and full release workflows remain. |

### Current behavior and settled implementation choices

- The UI uses the requested field-guide design: warm paper light mode and warm
  charcoal dark mode, with a local browser **Light / Dark / Auto** preference.
  This supersedes the original navy palette described in Phase 4.
- The interface is native HTML/CSS with one embedded theme script. No third-party
  JavaScript is required yet; HTMX/Alpine have not been vendored.
- Question browsing keeps its unsaved “I got this right” self-check. Practice
  tests have advisory automatic grading and save answer history locally, as do
  Flashcard reviews.
- All eight changing questions display the permanent USCIS verification link and
  the original source instructions. Four federal officeholders, state senators,
  and House representatives are bundled as a dated snapshot; the four federal
  offices and the learner's state governor are additionally refreshable from
  Wikidata. `--offline` disables the refresh action; it does not (and does not
  need to) touch the sidecar file read, which is local disk, not network.
- A dated 2026-09-21 bundled snapshot resolves President, Vice President,
  Speaker of the House, and Chief Justice, with its official source URL shown
  alongside the permanent USCIS verification banner. A local state selection
  resolves the state-capital question and displays both current senators, any
  one of whom grades as correct for Q23. The resolver enforces manual
  override, local sidecar, then bundled-data precedence, for every one of the
  eight changing questions. Settings saves a ZIP locally, shows every
  crosswalk candidate, and requires a district choice when more than one
  applies — or, once that many-candidate case shows up, offers a **Look up my
  district** address form as an alternative: it calls the Census Bureau's
  keyless geocoder, never stores the address itself, and keeps only the
  resulting district. Once state and district are set (by either path), Q29
  resolves automatically from the bundled House roster (falling back to a
  state's sole seat for at-large states and D.C.'s delegate, since the
  crosswalk's district numbering for those seats doesn't always match
  congress-legislators' own); Settings also provides a local manual
  representative override, paired with the official House lookup, that always
  takes precedence, for the day a seat turns over between rebuilds. A Settings
  **Refresh current officials**
  action fetches the four federal offices, plus the governor of the learner's
  selected state (Q61), from Wikidata through the bounded, fixture-tested
  `FederalClient`/`GovernorClient`, and writes them to a local
  `officials-live.json` sidecar (never overwriting the bundled snapshot),
  labeled with its source and fetch date wherever it resolves an answer; a
  **Reset to bundled** action removes the sidecar. A failed refresh renders the
  Settings page inline with a clear error and leaves any previous sidecar
  value untouched (including a previously refreshed governor for a *different*
  state — refreshing accumulates one governor per state visited, it does not
  replace the whole set), and the action is unavailable outright when the app
  is started with `--offline`. Unlike the original Phase 2.5 plan, governors
  are **not** bundled into `officials.json` as a static 50-state snapshot —
  fetching one unverified name per state in bulk and shipping it as if it were
  ground truth was judged riskier than resolving a single state live, on
  demand, through the same source-and-date-labeled, resettable path already
  built for the four federal offices. D.C. is excluded by construction: its
  `states.json` entry carries no Wikidata QID (validated at load), so a
  refresh silently skips the governor fetch for it, consistent with the
  question's own guidance that D.C. has no governor. Both Wikidata clients
  send an identifying `User-Agent` (required by Wikidata's usage policy;
  requests without one are rejected with HTTP 403) and decode only the fields
  they need from the real SPARQL JSON response — real responses nest
  additional standard fields (a top-level `head`, `type`/`xml:lang` beside
  every `value`) that a naive `DisallowUnknownFields` decode had rejected
  outright before this was caught by live-network testing, not by the
  existing fixture tests, which had (incorrectly) omitted those fields.
- Chapter content uses strictly decoded JSON front matter (a YAML subset) and a
  documented, limited Markdown syntax, rendered through escaped Go templates.
  Chapter links are authored once and reverse links are derived at startup.
- Chapters are labeled **text editions**. Prose, objectives, sidebar text,
  diagram/map labels, and printed captions are retained; map geography and
  uncredited illustrations are not reproduced. This does not yet satisfy the
  full visual-content scope of the release plan. Credited images retain source
  resolution; downscaling has not been implemented.
- As of Chapters 1, 6, and 7, genuine geography is illustrated with real,
  individually-verified public-domain maps sourced from outside the four
  supplied PDFs (National Archives, USGS) — a deliberate, narrow exception to
  "the PDFs are the source of truth," made only where the Study Guide's own
  map graphic has no printed credit, an equivalent government map exists and
  has been personally verified (not a generic "fair use" search result), and
  the map is added alongside rather than replacing the exact-fact text
  transcription (dates, territory names) it cannot itself convey. See
  `internal/content/data/images/EXTERNAL-SOURCES.txt` for the exact source
  URL, quoted license statement, and verification date for each. Custom
  textbook diagrams (branch trees, flowcharts, the Cabinet-position and
  naturalization-eligibility lists) have no such equivalent and remain text.
- Runtime validation checks questions, chapter references, image captions and
  credits, approved image files, and bidirectional links. Full 128-question
  chapter coverage remains a release gate once all chapters are authored.

### Verification and known limits

- Passing at the Chapter 1–10 milestone: `go test -race ./...`, `go vet ./...`, native
  build, source word coverage by page, chapter/reader goldens, and route/image
  tests. The initial implementation also passed all six cross-platform builds.
- Chromium checks passed for theme switching/persistence, system appearance,
  chapter contents anchors, four images, question backlinks, and no horizontal
  overflow at 380px. Automatic browser launch has not been manually checked.
  Chapters 3–10's content was verified by direct visual PDF review instead
  (below).
- `make lint` is **not fully verified**: the available Staticcheck revision cannot
  read this machine's Go 1.27 export format. Vet passes; Staticcheck needs a
  compatible toolchain/tool version.
- Source word inventories detect omissions/additions, not reading order. Chapter 2
  was also visually reviewed page-by-page against the Study Guide PDF (pages
  18–23): prose, the two "U.S. Congress" tree diagrams, the lawmaking flowchart,
  the credited Oval Office photo, and the blank page 23 all match the source.
  The uncredited Capitol photo on page 20 is omitted per the image-review policy;
  its printed caption is retained as reader text. Chapter 2's page 22 lawmaking
  diagram has no extractable text in the PDF; its visually transcribed labels
  are recorded in `data/raw/study-guide-visual-supplement.json`.
- Chapter 3 (pages 24–28) was visually reviewed the same way: prose, both
  three-branch diagrams, the full 22-item Cabinet-level list, and the Line of
  Succession diagram all match the source, and both have extractable PDF text
  (no visual supplement entry needed). Of the seven images on these pages, only
  the 1964 voting photo (page 26) carries a printed institutional credit
  (Library of Congress); the rest — the 2025 Cabinet photo, both presidential
  portraits, the FBI/USCIS and armed-forces seals, and the party icons — are
  omitted as uncredited, with printed captions kept as reader text where any
  exist. Omitting the sitting President's photo also avoids the same shelf-life
  problem as the eight changing answers.
- Chapter 4 (pages 29–32) was visually reviewed the same way: prose, both
  diagrams (the repeated three-branch tree and the Federal Court System
  pyramid), the Chief Justice duties list, and the majority sidebar all match
  the source, with extractable PDF text throughout. None of its five
  photographs carry a printed credit, so none are reproduced. This chapter
  exposed a real bug in the source-coverage test: its page 31 repeats the
  chapter title verbatim as a genuine section heading, and a stray "CHAPTER 1:
  THE U.S. CONSTITUTION" running header bleeds onto its page 29. The test's
  page-furniture filter was generalized (any `CHAPTER N` banner line, and the
  bare-title match restricted to the chapter's opening page only) rather than
  hard-coding this chapter's two exceptions; re-run against Chapters 1–3
  confirmed no behavior change for their existing goldens.
- Chapter 5 (Rights and Responsibilities, pages 33–37) was visually reviewed
  the same way and caught a real transcription error before commit: a first
  draft dropped "in federal elections" from one sentence, which the coverage
  test flagged immediately (a concrete demonstration of why the test exists).
  Of its five photographs, four have no printed credit and one (the 2008
  voting-booth photo) is credited to a non-federal source, "Courtesy of the
  Polling Place Photo Project"; none are reproduced pending a deliberate
  licensing review of that credit, distinct from the no-credit-at-all
  omissions elsewhere. One sentence about the Federalist Papers splits across
  pages 33/34 in the source itself, reproduced as a matching mid-sentence
  block split. Two source wording issues ("Civils Rights Act of 1964"; "will
  raise their right hand say the Oath of Allegiance", missing "and") are kept
  exactly as printed.
- Chapter 6 (U.S. Geography, pages 38–42) was visually reviewed the same way
  and surfaced the most PDF-extraction quirks of any chapter so far, all
  resolved by cross-checking the rendered page images rather than guessed:
  three map labels that follow a curved/diagonal path (page 39's "Washington,
  D.C." marker, page 40's "GULF OF AMERICA", page 41's mountain/river names)
  extract as scrambled letter fragments, handled via the visual-supplement
  mechanism; page 41 independently duplicates "Alaska", "Hawaii", and all five
  territory names in a hidden ALL-CAPS text layer behind the visible
  mixed-case labels, confirmed by inspection to have only one printed instance
  each, so the coverage test drops one duplicate of each; and this chapter's
  running header uses an en dash instead of the colon Chapters 1–5 used,
  requiring the banner filter to be broadened (re-run against Chapters 1–5
  confirmed no change to their goldens). None of Chapter 6's five main map
  graphics are reproduced as images: none carry a printed credit, and the
  full state/territory geography is not reproduced, consistent with Chapter
  1's map policy.
- Chapter 7 (Early American History, pages 43–46) was visually reviewed the
  same way: prose, both extractable maps (the transatlantic-routes map and
  the 13-colonies map), and the sidebar on the 1707 union of England and
  Scotland all match the source. Two of its four photographs are reproduced:
  a Jamestown street scene credited to the National Park Service — a federal
  agency, reviewed and added to the credit allowlist, the first addition since
  the JFK Library in Chapter 2 — and a painting already covered by the
  existing Library of Congress credit. The other two (an illustration credited
  to the Jamestown Yorktown Foundation, a Virginia state institution, and an
  uncredited illustration of the first enslaved people arriving in Jamestown)
  are omitted, the former pending the same kind of deliberate licensing review
  as Chapter 5's Polling Place Photo Project credit. Page 45's "Atlantic
  Ocean" map label follows the ocean's curve and extracts as scrambled
  fragments, handled via the visual-supplement mechanism like several of
  Chapter 6's labels. This chapter covers Native American depopulation and
  the start of slavery in the colonies factually, in the source's own words.
- Following an explicit decision to source real geography images rather than
  rely solely on text transcriptions, Chapters 1, 6, and 7 were retrofitted
  with two verified public-domain reference maps (four manifest entries: the
  same National Archives map is used at three different chapter pages, so it
  is copied under three separate image IDs — catalog.go ties one manifest
  image to exactly one page). Each candidate was found via search, then its
  license statement was independently confirmed on the actual source page
  before download; a "fair use" internet image search was explicitly
  rejected as the sourcing method. Two categories (a territories-with-capitals
  map; a historical map of both colonization and slave-trade Atlantic routes)
  had no verifiable public-domain candidate and remain text. Library of
  Congress item pages returned HTTP 403 to automated fetching throughout this
  research, ruling out several otherwise-promising candidates. Both added
  images are supplementary to, not replacements for, the exact facts (dates,
  territory names) their chapters' text transcriptions already convey.
- Chapter 8 (The American Revolutionary War & The Declaration of Independence,
  pages 47–51) was visually reviewed the same way and is the richest image
  chapter yet: all six of its Study Guide photographs/paintings are
  reproduced, none omitted for lack of a credit. Three credits were already
  allowlisted (National Archives, National Park Service, Library of
  Congress); "Courtesy of the U.S. Senate." is a new addition, reviewed as a
  federal legislative body's own art collection. Only the uncredited
  Benjamin Franklin portrait is omitted. The Currier & Ives print's credit
  line carries a National Archives item number ("...National Archives,
  532915."); it is moved into the caption so the credit field holds the
  exact allowlisted string, losing no text. A sentence spans the page 47/48
  boundary in the source itself, handled with the same mid-sentence block
  split as Chapter 5. This chapter's printed title splash spans two lines
  rather than one, which no prior chapter's did; the coverage test's
  title-splash filter was generalized to match the full splash span's
  concatenation rather than a single line, re-confirmed against Chapters
  1–7's existing goldens. The National Archives' 1775 Mitchell Map is reused
  here too (a fourth manifest copy), alongside this chapter's own 13-states
  content.
- Chapter 9 (A New Government and an Expanding Nation, pages 52–57) was
  visually reviewed the same way. Only one of its six illustrations carries
  a printed credit (the Constitution document, already allowlisted); the
  other five have none and are omitted, captions retained as text. Unlike
  every image chapter since Chapter 1, none of its three map graphics have
  any extractable text layer beyond their captions, so no visual-supplement
  entry was needed for this chapter. This chapter substantially repeats
  content already covered in Chapters 1, 5, 6, and 8 (the Constitutional
  Convention, the Federalist Papers, George Washington's "Father of Our
  Country" epithet) with new specifics (exact presidential dates); genuinely
  repeated facts are linked to every chapter that states them, consistent
  with this project's established practice.
- Chapter 10 (The Civil War, pages 58–61) was visually reviewed the same way.
  Four of its eight illustrations carry a printed credit and are reproduced
  (cotton picking and Susan B. Anthony from the Library of Congress,
  Frederick Douglass from the National Park Service, the Nast "Emancipation"
  print from the Library of Congress); the other four have none and are
  omitted. The "Emancipation" print's own credit line is missing the
  trailing period this chapter's other two Library of Congress credits use;
  the manifest's credit field holds the standard normalized string, and the
  punctuation difference is noted rather than silently carried through. As
  with Chapter 9, none of this chapter's uncredited illustrations have any
  extractable text layer beyond their captions. Frederick Douglass has no
  dedicated question in the official 128 and is covered anyway, per this
  project's second goal of teaching the material, not only the test answers.
- Chapter 11 (American History: 1900-2001, pages 62–68) was visually
  reviewed the same way. Ten of its fifteen illustrations carry a printed
  credit and are reproduced; the other five (the Cold War map, Dr. Martin
  Luther King Jr., Dolores Huerta, the March on Washington, and the
  September 11 rescue-workers photo) have none and are omitted. Two credit
  sources are new: "Courtesy of NASA." (the Apollo 11 Moon-landing photo)
  and "Courtesy of the White House." (the Pentagon flag photo). The Pentagon
  flag photo's printed credit reads "White House photo by Paul Morse," with
  no "Courtesy of..." phrasing in the source at all; normalizing it to the
  allowlisted string introduces three words not literally printed, recorded
  in the visual supplement for page 67. The Cold War map's "United States"
  and "Soviet Union" labels are cleanly extractable and transcribed as a
  diagram; its "Atlantic Ocean" label follows the ocean's curve and extracts
  as scrambled fragments, handled the same way as several earlier chapters'
  map labels. This chapter restates two facts already covered elsewhere with
  new specifics that justify new links: the federal government's exclusive
  power to declare war (Chapter 1 states it generally; this chapter narrates
  Congress exercising it in 1941) and the 22nd Amendment's two-term limit
  (naming the amendment directly, unlike Chapter 3's mention of the rule
  without naming its source).
- Chapter 12 (American Symbols and Holidays, pages 69–76) was visually
  reviewed the same way, completing all 12 chapters. Three of its fifteen
  captioned illustrations carry a printed credit and are reproduced (the
  Statue of Liberty and Abraham Lincoln from the Library of Congress,
  "George Washington at Princeton" from the U.S. Senate); one more carries a
  printed credit that is explicitly excluded, "Raising the Flag on Iwo
  Jima," credited to "the Associated Press" — the first chapter where the
  project's standing AP/wire-service rejection rule actually excludes an
  image. The other eleven captioned illustrations have no printed credit
  and are omitted. Two further graphics (the chapter-opening banner and the
  full flag illustration) have no caption at all and are not counted as
  illustrations, the same treatment given every chapter's uncaptioned
  opening banner. This chapter adds five new multi-chapter links: George
  Washington's "Father of Our Country" epithet (Q86, now four chapters),
  the War of 1812 and Civil War as 1800s wars (Q91), the 50 stars (Q122),
  Veterans Day (Q128), and national holidays generally (Q126, now three
  chapters), for which this chapter's own "U.S. Holidays" list is the
  definitive source. Q125 ("What is Independence Day?") is a new link,
  stated directly by this chapter's Independence Day section. Q8 is
  deliberately not linked here despite similar phrasing, for consistency
  with Chapter 8's existing scope.
- Grading now takes precedence over the remaining library. Its test contract is:
  every fixed official bullet must match **as an item**, while a question only
  passes when its required distinct-item count is met. “Answers will vary” and
  “Visit…” are instructions rather than factual answers, so the first grading
  engine reports the eight changing questions as unavailable until the officials
  overlay supplies a resolved value; the user’s self-check remains authoritative.

### Source corrections to the planning notes

The supplied Q&A PDF has **22** Q48 bullets, **six** Q67 bullets, and section
boundaries Q16–62 (System of Government), Q63–72 (Rights and Responsibilities).
The original planning counts/ranges below are corrected accordingly. Required
answer counts are unchanged. Source typos and source-specific wording are retained
and flagged in chapter notes, not silently corrected.

### Next work

1. All 12 chapters are authored. 107 of 128 official questions are linked to
   a chapter from prose that literally states an accepted answer; the other
   21 are not stated verbatim anywhere in the 12 chapters and remain
   deliberately unlinked — full 128-question coverage was never assumed to
   follow automatically from finishing the chapters, and isn't a
   near-term goal. Separately, decide whether to pursue a licensing review
   for the non-federal photo credits deferred so far: the Polling Place
   Photo Project (Chapter 5) and the Jamestown Yorktown Foundation
   (Chapter 7); neither is blocking.
2. Officials resolution is now complete: sidecar persistence, the Settings
   refresh action, the `--offline` guard, a live per-state governor refresh
   (Q61), the bundled House roster (Q29), and the Census address geocoder
   (an alternative to the ZIP candidate-list picker for a multi-district ZIP)
   are all done. All eight changing questions resolve automatically wherever
   their source data allows it, each labeled with its source and date,
   through the same manual > sidecar > bundled precedence.
3. The Citizen's Almanac library documents are underway: "Patriotic Anthems
   of the United States" (national anthem, "America the Beautiful," "The New
   Colossus") is published, source-verified against the rendered PDF pages
   the same way every chapter has been, with a new `almanac-visual-supplement.json`
   for its non-extractable decorative titles and an eight-page front-matter
   offset documented in data/library/README.txt. Remaining: "Patriotic
   Symbols of the United States" (Pledge of Allegiance, Flag, Motto, Great
   Seal — pages 20-26, already page-verified during this session), the seven
   speeches, and the four landmark Supreme Court cases, each carrying the
   required citation.

---

## Context

Goal: an application to prepare for the USCIS civics test (part of the N-400
naturalization application) that does two jobs at once:

1. **Pass the test.** Drill the official 128 questions until the answers are automatic.
2. **Actually learn.** Go past flashcard level into a real guide to American
   government and history, so the material means something rather than being
   memorized noise.

The project began with four PDFs and no code. This plan targets a **single Go binary**
serving a modern, responsive web UI with four sections — **Learn**, **Library**,
**Flashcards**, **Practice Test** — all content derived from the official PDFs,
working fully offline, with tests that verify the content against its sources.

### Decisions made

| Decision | Choice |
|---|---|
| Frontend | Go `html/template` + vendored HTMX/Alpine + hand-written CSS. **No Node/npm.** `go build` is the entire build. |
| Changing answers | First-run wizard collects **state + ZIP**. Bundled dated snapshots work offline; an on-demand refresh pulls current officeholders from keyless authoritative sources. Verification banner always shown. |
| Images | Study Guide only, gov-source only. AP-credited image dropped. **No Citizen's Almanac images** (its license forbids reuse outside that publication). |
| Content scope | 12 chapters + full reference library (Declaration, Constitution, 27 amendments, 7 speeches, symbols, 4 landmark cases). |

---

## The answer model (the subtlest part of this app)

A civics answer is **not** a string. Each question carries a list of
*acceptable answers*, and the question's wording says **how many distinct ones
you must supply**. Getting this wrong means the app marks correct answers wrong.

There are three distinct shapes, all verified against the PDF:

**1. Variants — supply any one (`required_count: 1`).** The bullets are
alternative phrasings or alternative valid facts. Any single one is a complete,
fully correct answer.

```
2. What is the supreme law of the land? *
     • (U.S.) Constitution                    <- one bullet, one answer

41. Name one power of the president.
     • Signs bills into law                   <- six bullets; ANY ONE is correct
     • Vetoes bills
     • Enforces laws
     • Commander in Chief (of the military)
     • Chief diplomat
     • Appoints federal judges
```

**2. Enumerations — supply N distinct ones (`required_count: 2, 3, or 5`).**
The bullets are a menu, and you must name several *different* items from it.

| Q | Question | Required | Bullets offered |
|---|---|---|---|
| 10 | Name **two** important ideas from the Declaration of Independence and the U.S. Constitution. | 2 | 6 |
| 48 | What are **two** Cabinet-level positions? | 2 | 22 |
| 65 | What are **three** rights of everyone living in the United States? | 3 | 6 |
| 67 | Name **two** promises that new citizens make in the Oath of Allegiance. | 2 | 6 |
| 69 | What are **two** examples of civic participation in the United States? | 2 | 10 |
| 81 | There were 13 original states. Name **five**. | 5 | 13 |
| 126 | Name **three** national U.S. holidays. | 3 | 11 |

**3. The trap: wording lies about cardinality.** A number in the question does
**not** imply `required_count > 1`. These all say "two" or "three" but a *single*
bullet is the whole answer:

```
16. Name the three branches of government.
     • Legislative, executive, and judicial   <- ONE bullet answers it
     • Congress, president, and the courts    <- or this one; they're variants

19. What are the two parts of the U.S. Congress?
     • Senate and House (of Representatives)  <- ONE bullet

15. There are three branches of government. Why?
     • So one part does not become too powerful
     • Checks and balances                    <- variants, any one
     • Separation of powers
```

**Therefore `required_count` cannot be derived by regex and must be authored
per question, then validated** (see Phase 2.6). The seven questions in the table
above are the complete `required_count > 1` set; every other question is 1.

Two further wrinkles inside individual answer strings:

- **Parentheses mark optional text.** `(U.S.) Constitution` must accept both
  "Constitution" and "U.S. Constitution". `Serve (help, do important work for)
  the nation (if needed)` must accept the bare core.
- **Brackets are instructions to the reader, not answers.** `Answers will vary.
  [District of Columbia residents should answer that D.C. does not have a
  governor.]` — the bracketed text is guidance and must never be graded against.

### Data model

```go
type Question struct {
    ID            int
    Section       string   // "American Government"
    Subsection    string   // "A: Principles of American Government"
    Prompt        string
    Answers       []Answer
    RequiredCount int      // 1 for most; 2, 3 or 5 for the seven enumerations
    Is6520        bool     // one of the 20 asterisk-marked questions
    AnswerKind    Kind     // Fixed | StateSpecific | CurrentOfficial
    Chapters      []string // authored in Phase 2
}

type Answer struct {
    Text     string // verbatim from the PDF, e.g. "(U.S.) Constitution"
    Core     string // required portion,  e.g. "Constitution"
    Full     string // with optionals,    e.g. "U.S. Constitution"
    Guidance string // bracketed reader instruction; never graded
}
```

`RequiredCount == 1` means *any one* of `Answers` is a complete response.
`RequiredCount == n` means the user must supply *n distinct* members of `Answers`.

---

## Answers that change: onboarding, lookup and refresh

**Eight questions** have no fixed answer. Four are also 65/20 questions (marked ★),
so even the shortest study path hits them:

| Q | Question | Varies by | Source |
|---|---|---|---|
| 23 | Who is one of your state's U.S. senators now? | state | congress-legislators |
| 29 | Name your U.S. representative. | **district** (ZIP) | crosswalk + congress-legislators |
| 30 ★ | Speaker of the House now? | time | Wikidata |
| 38 ★ | President now? | time | Wikidata |
| 39 ★ | Vice President now? | time | Wikidata |
| 57 | Chief Justice now? | time | Wikidata |
| 61 ★ | Governor of your state now? | state + time | Wikidata |
| 62 | Capital of your state? | state | bundled static table |

### One honest caveat about "web search"

A Go binary cannot do open-ended web search the way a chat assistant can — and
for *test answers* that is a feature, not a limitation. Search-engine snippets
are exactly the wrong input for a fact that must be correct on interview day.
So this is built as **targeted lookups against keyless, authoritative, structured
sources** rather than a search engine: no API key, no scraping, no snippet
parsing, no per-user cost, and a stable schema that can be tested against
recorded fixtures. All three sources below were verified to respond correctly
during planning.

### Sources (all keyless, all verified)

**1. ZIP -> congressional district — bundled, offline.**
Census `tab20_cd11920_zcta520_natl.txt` (119th Congress <-> 2020 ZCTA relationship
file), pipe-delimited, fetched once at ingest and trimmed to two columns
(`GEOID_ZCTA5_20`, `GEOID_CD119_20`).

    Measured: 40,147 pairs / 33,791 distinct ZIPs -> 432 KB raw, 88 KB gzipped.

Small enough to embed outright, so district lookup needs **no network at all**.
Source: `https://www2.census.gov/geo/docs/maps-data/data/rel2020/cd-sld/tab20_cd11920_zcta520_natl.txt`

> **ZIPs are not districts.** Verified in the data: some ZIPs span up to **four**
> congressional districts (90002, 90022, 90640, 92324, 93550 among them). When a
> ZIP maps to more than one district the wizard lists every candidate
> representative and asks the user to pick — it never guesses. A user who is
> unsure can enter a full street address and take the exact path below.

**2. Exact district from a street address — optional, online.**
Census Geocoder, keyless, returns a `119th Congressional Districts` geography
layer (verified). Used only when the user opts into address entry to disambiguate
a multi-district ZIP. The address is sent to the Census Bureau and **never
stored** — the app keeps only the resulting district number.

**3. Senators and representatives.**
`unitedstates/congress-legislators` -> `legislators-current.json` (verified 200,
~1.4 MB). Well-maintained public dataset keyed by state and district.

**4. President / VP / Speaker / Chief Justice / governors.**
Wikidata SPARQL, one query, no key. Offices by QID via `P1308` (officeholder);
governors via `P6` on the state entity. Verified live during planning — a single
query returned all four federal officeholders correctly, and `wd:Q1439 wdt:P6`
returned the sitting governor of Texas.

| Office | QID | Property |
|---|---|---|
| President | `Q11696` | `P1308` |
| Vice President | `Q11699` | `P1308` |
| Speaker of the House | `Q912994` | `P1308` |
| Chief Justice | `Q11147` | `P1308` |
| Governor of <state> | state QID | `P6` |

### Offline-first, network-optional

The binary ships with a **dated snapshot** of every value above. The app is fully
usable with the network unplugged, forever. Network is strictly an enhancement:

- **Startup never blocks on network.** No lookup on the hot path, ever.
- The app knows the **exact date its data expires** (see "Staying current"
  below) and warns ahead of it. It does **not** auto-fetch.
- **Refresh is user-initiated** (a button in Settings, and on the prompt). It runs
  with a short timeout against all sources concurrently, and **any failure is
  non-fatal** — a source that times out keeps its previous value and reports so.
- Fetched values are written to `~/.config/n400/officials-live.json`, which
  overlays but never overwrites the embedded snapshot. "Reset to bundled" always
  works.
- Every fetched answer is labeled with its **source and fetch date** in the UI.
- **A manual override always wins.** Any of the eight can be typed in by hand and
  that value is treated as authoritative — the escape hatch for the day a source
  is wrong or lags an appointment.
- `--offline` disables all network code paths outright.

### The rule that does not bend

**The `uscis.gov/citizenship/testupdates` banner stays on all eight questions
regardless of where the answer came from — bundled, fetched, or hand-entered.**
None of these sources is USCIS. They are good enough to study against and not
good enough to stake an interview on, and the app must never blur that line. The
PDF's own instruction is reproduced verbatim alongside the value.

---

## Staying current (new Congress, elections, appointments)

The app is offline-first, so "how does it update?" needs a real answer rather
than a staleness guess. Three things change on three different clocks, and they
fail differently — the design treats them differently.

### The data tells us when it expires

`legislators-current.json` carries an explicit `end` date on every term. Checked
against live data during planning:

| Term end | Count | Meaning |
|---|---|---|
| 2026-11-03 | 2 | appointed senators serving until a special election |
| **2027-01-03** | **472** | **all representatives + 1/3 of senators — the 120th Congress seats** |
| 2029-01-03 | 32 | senators, class continuing |
| 2031-01-03 | 33 | senators, class continuing |

So the app does not guess. At ingest, `cmd/ingest` records
`expires_on = min(term.end)` across the bundled set, and the app computes a
**hard expiry date** from the data itself. The UI can say the true thing —
*"Your representative and senator data is current through January 3, 2027, when
the 120th Congress is seated"* — and start prompting ahead of that date rather
than after the answers have quietly gone wrong.

### What changes, when, and how badly it breaks

| Data | Changes | On a stale copy |
|---|---|---|
| President / VP | Jan 20 after a presidential election | **Wrong immediately** |
| Speaker | New Congress, or a mid-term ouster (2023 precedent) | **Wrong immediately**, unpredictably |
| Representatives, ~1/3 of Senate | Jan 3 of odd years | **Wrong immediately** |
| Chief Justice, appointed senators | Irregular (death, retirement, resignation) | Wrong without warning |
| Governors | Nov even years; VA/NJ/KY/LA/MS odd years | Wrong immediately |
| **District boundaries** (crosswalk) | Redistricting: post-census (2032) + mid-decade court orders | **Usually still correct** — degrades gently |
| State capitals | Effectively never | Fine |
| USCIS questions / chapters | When USCIS revises the test | Needs a rebuild, not a data refresh |

That sixth row is the useful nuance: a stale *crosswalk* is mostly harmless
because district lines only move on redistricting, while a stale *officeholder
name* is wrong the morning after a transition. The app's warnings are
proportionate to that — it does not nag about geography the way it nags about
names.

### Three update paths, in order of user effort

**1. Refresh button — covers every officeholder, takes a second.**
Settings -> *Refresh current officials*. Online, user-initiated, concurrent,
short timeout, per-source failures non-fatal. Writes `officials-live.json` to
the config dir as an overlay. This is the normal answer for a new Congress: on
January 4, 2027, press the button.

**2. Sidecar file — no network, no new binary, no infrastructure.**
Drop a JSON file in `~/.config/n400/` and it overlays the embedded data on next
start. Same schema as the bundled snapshot, validated on load with clear errors.
This is the escape hatch for an air-gapped machine, a source that has gone away,
or a value the app got wrong. Documented in the README with a worked example.

**3. New binary — for content, not just data.**
A USCIS test revision, new chapters, or new images require re-ingesting the PDFs
and rebuilding. `make ingest && make build-all`. The app shows its content
version and build date in Settings so it is obvious when a rebuild is the answer.
The app does **not** self-update: silently swapping a study binary is worse than
telling someone to download one.

### Handling the Congress-number rollover

The crosswalk URL embeds the Congress number —
`.../rel2020/cd-sld/tab20_cd{N}20_zcta520_natl.txt` — so it moves from `cd119`
to `cd120` in January 2027. Verified during planning: `cd118` and `cd119` both
return HTTP 200, the pattern is stable, and **`cd120` is currently 404** —
Census publishes the new file some time after the Congress is seated.

`cmd/ingest` therefore:

1. Derives the expected Congress number from the date —
   `N = (odd_year - 1789)/2 + 1`, seated Jan 3 of odd years. Verified:
   2025/2026 -> 119, 2027/2028 -> 120, 2033 -> 123.
2. Requests that file; **on 404, falls back to `N-1` and logs loudly** rather
   than failing the build. The 404 is expected for months after a new Congress,
   and the previous file's districts remain correct absent redistricting.
3. Records which Congress the bundled crosswalk actually came from, so Settings
   can display *"district map: 119th Congress"* honestly.

-> *verify:* Congress-number derivation is a pure function with table tests
across decade and century boundaries; the 404 fallback is tested against an
`httptest` server returning 404 for `cd120` and 200 for `cd119`, asserting the
ingest succeeds with the older file and records the older number.

### What the user actually experiences

- **Normal day:** nothing. No prompts, no network.
- **Approaching expiry** (data says Jan 3, 2027): a quiet, dismissible banner —
  *"A new Congress is seated Jan 3, 2027. Your representative and senator
  answers will need a refresh then."*
- **Past expiry, online:** the banner becomes a one-click refresh.
- **Past expiry, offline:** affected answers are flagged individually as
  *"may have changed — verify at uscis.gov/citizenship/testupdates"*. Everything
  else in the app — 120 of 128 questions, all chapters, the whole library —
  is unaffected and stays fully usable.

The eight changing questions are the only part of this app with a shelf life.
The other 120 questions, the 12 chapters, and the entire reference library never
go stale.

---

## What the source PDFs actually contain (verified)

**`2025-Civics-Test-128-Questions-and-Answers.pdf`** (19pp) — ground truth.
Text layer is clean and machine-parseable:

- Questions numbered `1.`–`128.`, **no gaps** (verified).
- Answers as `• ` bullets under each question.
- Three sections / eight subsections:
  - `AMERICAN GOVERNMENT` -> `A: Principles of American Government` (1–15),
    `B: System of Government` (16–62), `C: Rights and Responsibilities` (63–72)
  - `AMERICAN HISTORY` -> `A: Colonial Period and Independence`, `B: 1800s`,
    `C: Recent American History and Other Important Historical Information`
  - `SYMBOLS AND HOLIDAYS` -> `A: Symbols`, `B: Holidays`
- **Exactly 20** asterisk-marked (65/20) questions, verified:
  `2, 7, 12, 20, 30, 36, 38, 39, 44, 52, 61, 66, 74, 78, 86, 94, 113, 115, 121, 126`.
- **Variable answers**: `Answers will vary` (Q23 senator, Q29 representative,
  governor, state capital) and `Visit uscis.gov/citizenship/testupdates`
  (Q30 Speaker, Q38 President, Q39 VP, Chief Justice).
- Test rules: officer asks up to 20 of 128, **pass at 12**. 65/20 applicants
  are asked 10 of the 20 starred, **pass at 6**.

**`USCIS-2025-Civics-Test-Study-Guide.pdf`** (88pp, 40MB) — *"One Nation, One
People: The USCIS Civics Test Textbook"*. 12 chapters, each opening with a
`❏`-bulleted "In this chapter, you will learn about:" list. Body page numbers
match PDF page numbers (verified at Ch. 2). 110 embedded images with printed
credit captions.

| Ch | Title | Pages |
|---|---|---|
| 1 | The U.S. Constitution | 8–17 |
| 2 | The Legislative Branch | 18–23 |
| 3 | The Executive Branch | 24–28 |
| 4 | The Judicial Branch | 29–32 |
| 5 | Rights and Responsibilities | 33–37 |
| 6 | U.S. Geography | 38–42 |
| 7 | Early American History | 43–46 |
| 8 | The American Revolutionary War and the Declaration of Independence | 47–51 |
| 9 | A New Government and an Expanding Nation | 52–57 |
| 10 | The Civil War | 58–61 |
| 11 | American History: 1900–2001 | 62–68 |
| 12 | American Symbols and Holidays | 69–76 |
| — | Index: 128 Questions and Answers | 77–88 |

> **Important gap:** the Study Guide contains **no explicit question-number
> markers** in chapter bodies (verified by grep). The question->chapter mapping
> does not exist in any source and **must be authored by hand** (Phase 2.2).

**`DOI-Constitution-M-654.pdf`** (64pp) — Declaration of Independence +
Constitution. Clean structural markers: `Article. I.`–`Article. VII.`,
`Section. N.`, `Amendment I.`–`Amendment XXVII.`, with footnote digits appended
to amendment headings (`Amendment XIII.16`) that the parser must strip.

**`CitizensAlmanac-M-76.pdf`** (112pp) — text is public information and reusable
**with citation**; images are **not**. Usable: Rights/Responsibilities of a
Citizen; anthems & symbols (Star-Spangled Banner, America the Beautiful, The New
Colossus, Pledge, Flag, Motto, Great Seal); 7 speeches (Washington Farewell,
Lincoln 1st Inaugural, Gettysburg, FDR Four Freedoms, JFK Inaugural, MLK I Have
a Dream, Reagan Brandenburg Gate); founding documents; 4 landmark Supreme Court
cases (Marbury, Plessy dissent, Barnette, Brown); Prominent Foreign-Born Americans.

---

## Architecture

```
n400/
  cmd/
    n400/main.go            # the app: serve + open browser
    ingest/main.go          # dev-only: PDFs -> data/raw/*.txt + images (needs poppler)
  internal/
    content/
      questions.go          # Question/Answer models, loader
      parse_questions.go    # raw text -> []Question
      parse_answer.go       # "(U.S.) Constitution" -> Core/Full/Guidance
      chapters.go           # shared Markdown/JSON-front-matter parser + chapter loader
      library.go            # library document model + loader (Declaration, Constitution, Amendments, Almanac speeches/symbols/cases), same parser as chapters
      images.go             # image manifest loader
      validate.go           # cross-reference integrity checks
      data/                 # <-- go:embed root, all checked in
        raw/                #   pdftotext output (checked in; keeps tests hermetic)
        questions.json      #   generated, checked in
        required_counts.json#   authored cardinality overrides
        chapters/*.md       #   authored, front-mattered
        library/*.md        #   authored, front-mattered (same grammar as chapters)
        officials.json      #   dated snapshot (federal + governors)
        states.json         #   50 states + DC + territories, capitals, QIDs
        zcta_cd.txt.gz      #   embedded ZIP -> district crosswalk (88 KB)
        images/             #   curated JPEG/PNG + manifest.json
    quiz/
      engine.go             # selection, scoring, pass thresholds
      grade.go              # answer normalization + set matching
    flashcard/
      deck.go               # deck construction + filters
      scheduler.go          # Leitner boxes, injectable clock
    officials/
      snapshot.go           # load bundled + live overlay, staleness check
      district.go           # ZIP -> district(s) via embedded Census crosswalk
      fetch_congress.go     # congress-legislators client
      fetch_wikidata.go     # Wikidata SPARQL client
      fetch_geocoder.go     # Census geocoder (address -> district)
      resolve.go            # question ID -> displayed answer + provenance
    store/
      store.go              # progress persistence (JSON, atomic writes)
    web/
      server.go routes.go handlers_*.go
      templates/*.gohtml
      static/{app.css, htmx.min.js, alpine.min.js, fonts/}
```

**Runtime:** binds `127.0.0.1:8400` (`--port` to override), opens the default
browser (`--no-browser` to skip), serves everything from `embed.FS`. Fully
offline. Cross-compiles to linux/darwin/windows with no cgo.

**Progress storage:** one JSON file in `os.UserConfigDir()/n400/progress.json`,
written atomically (temp file + rename), behind a `store.Store` interface so
SQLite can replace it later if history outgrows it. Not now.

---

## Phase 1 — Skeleton and extraction pipeline

**1.1** `go mod init`; `cmd/n400/main.go` serving a page from `embed.FS`, graceful
shutdown, `--port` / `--no-browser`.
-> *verify:* `go build && ./n400` opens a browser to a working page.

**1.2** `cmd/ingest` — dev-only, shells out to `pdftotext -layout` and
`pdfimages -j`, writes `internal/content/data/raw/*.txt` and
`data/images/_extracted/`. Poppler is a **dev** dependency, documented in README.
-> *verify:* raw text files appear and are committed. **From here on tests never
touch the PDFs or need poppler** — they parse the checked-in raw text, keeping
CI hermetic and fast.

**1.3** `parse_answer.go` — split one bullet into `Core` / `Full` / `Guidance`:
strip `(optional)` parens into Core-vs-Full, lift `[bracketed]` text into
Guidance, preserve `Text` verbatim.
-> *verify:* table tests over every real-world shape in the PDF —
`(U.S.) Constitution`, `Serve (help, do important work for) the nation (if needed)`,
`Presidents Day (Washington's Birthday)`, `Secretary of War (Defense)`,
`Answers will vary. [District of Columbia residents...]`, and plain strings.

**1.4** `parse_questions.go` — raw text -> `[]Question`. Must handle: page-break
furniture (`N of 19`, `uscis.gov/citizenship`), the asterisk both trailing a
prompt and isolated on its own line (1 occurrence), prompts wrapping across
lines, answers wrapping across lines, and bracketed clarifications.
-> *verify (tests):* exactly 128 questions; IDs 1–128 contiguous; exactly 20
`Is6520` and the ID set equals the verified list; every question has >= 1 answer;
every Section/Subsection in the known set; **golden-file test** pinning the full
parse so future parser edits cannot silently corrupt content.

---

## Phase 2 — Content authoring (the long pole)

**2.1 Chapters** — 12 Markdown files, e.g. `data/chapters/01-constitution.md`:

```yaml
---
id: constitution
number: 1
title: The U.S. Constitution
objectives:          # from the ❏ list on the chapter's first page
  - The U.S. Constitution.
  - When the Constitution was written.
questions: [1, 2, 3, 4, 5, 6, 7, 10, 14, 15]
images: [ch01-christy-signing, ch01-constitution-archives]
source_pages: "8-17"
---
Body prose, reflowed from the PDF's two-column layout...
```

Prose is transcribed faithfully from the Study Guide — no paraphrasing, no
invention. The two-column layout means `pdftotext -layout` output needs manual
reflow per chapter.

**2.2 Question -> chapter mapping** — the hand-authored link no source provides.
Each of the 128 questions maps to >= 1 chapter; the reverse index powers
"questions covered in this chapter."
-> *verify:* every question ID appears in >= 1 chapter list; every chapter has
>= 1 question; no chapter references a nonexistent ID; union covers all 128.

**2.3 Library** — `data/library/*.md`, hand-authored and source-verified the
same way as chapters (visual PDF review, not a generic layout parser), reusing
chapters' exact Markdown grammar and JSON front matter via a shared parser
(`splitDocument`/`parseBlocks` in `chapters.go`). Each document carries `kind`
(`founding-document` | `speech` | `symbol` | `case`), `source`
(`declaration-constitution` | `citizens-almanac`), and, for Almanac-sourced
documents, the exact required `citation` string — validated by
`ParseLibraryDoc`, not just documented. Three documents are authored so far,
all from `DOI-Constitution-M-654.pdf`, single-column and comparatively easy
to verify against the PDF unlike the Study Guide's two-column chapters:
`declaration.md` (pages 7–13, including its signature block and signers list
as their own short blocks/lists rather than run into prose) and
`us-constitution.md` (pages 15–38: the Preamble, Articles I–VII with
Sections, the Constitution's own signers list, and — for completeness — the
two procedural pages the source prints immediately after it and before the
Amendments, the ratifying convention's closing resolution and the first
Congress's resolution transmitting the Bill of Rights). Several sentences
that split across a page boundary in the source are authored as matching
split blocks in both documents, the same treatment chapters give a
mid-sentence page break. `us-constitution.md` also inlines the source's 11
numbered footnotes (each marking text later superseded by a specific
amendment) as a parenthetical at their point of reference — e.g. "for six
Years;" printed with "]2" in the source becomes "for six Years;" followed by
"(Changed by the Seventeenth Amendment.)" — rather than a separate
footnote/superscript UI; `TestLibrarySourceCoverage` strips the bare
reference digit from the raw side via a small, exact-match `footnoteRefs`
map (the same targeted-substitution technique chapters' `mapLabelFixes` uses
for PDF-extraction artifacts) so the word inventory still balances exactly.
`amendments.md` (pages 39–53) completes Amendments I–XXVII. The source's
heading footnote markers are stripped by the same exact-match coverage logic,
while every printed ratification note remains on its source page; the bracketed
Eighteenth Amendment and its repeal note are retained. It links 13 questions
whose accepted answers appear directly in the amendment text.

`patriotic-anthems.md` (pages 9–15) is the fourth document and the first
from `CitizensAlmanac-M-76.pdf` rather than the Constitution booklet. Its raw
extraction (`data/raw/almanac.txt`) splits on the PDF's own page breaks,
which include eight unnumbered/roman-numeral front-matter pages before the
printed Arabic page 1 begins; the document is authored with the printed page
numbers a reader sees in the book (matching the Constitution documents'
convention), so `TestLibrarySourceCoverage` applies an eight-page offset only
when indexing into the raw almanac pages for `citizens-almanac`-sourced
documents. It covers three of the nine items in the Almanac's own "Patriotic
Anthems and Symbols of the United States" section (pages 9–26) — the
national anthem, "America the Beautiful," and "The New Colossus," the three
the section's own intro names directly — leaving out two poems in the same
printed section (Whitman's "I Hear America Singing," Emerson's "Concord
Hymn," pages 16–19) that are outside this project's planned Almanac scope;
the section's other four items (Pledge, Flag, Motto, Great Seal, pages
20–26) are a separate document so both documents' page ranges stay fully
contiguous, since the coverage test requires every page within a document's
own range to be accounted for. Each song's decorative cursive title/author
attribution renders as pure vector art with no extractable PDF text layer at
all; its clean transcription lives in the new
`data/raw/almanac-visual-supplement.json`, the same mechanism the Study
Guide's non-extractable diagram/map labels already use, keyed by printed
page. The Almanac's two-column layout hyphenates words across a line wrap
far more often than the Constitution's single-column pages, and several
wraps land on a row that also carries the other column's unrelated text; a
new `almanacTextFixes` map fixes each affected region as one exact block,
the same targeted technique `footnoteRefs` uses. A new fenced ` ```verse `
block (parsed alongside chapters' existing ` ```text ` diagrams, rendered in
serif type rather than monospace) preserves each song's line breaks. It
links Q123 (the national anthem's name, stated directly); "America the
Beautiful" and "The New Colossus" have no dedicated question in the official
128 and are included anyway, per this project's second goal of teaching the
material, not only the test answers. Remaining: "Patriotic Symbols of the
United States" (Pledge, Flag, Motto, Great Seal), the Almanac's seven
speeches, and its four landmark cases, each carrying the required citation.

-> *verify:* `TestLibrarySourceCoverage` (a per-page word inventory against
`data/raw/constitution.txt`/`almanac.txt`, the same technique
`TestChapterSourceCoverage` uses for the Study Guide) and `TestLibraryGolden`/
`TestLibraryReferences` pass for every document; every cross-link resolves;
an Almanac-sourced document with a missing or altered citation fails to parse.

**2.4 Images** — curate the 110 extracted files down to keepers.
`data/images/manifest.json`:

```json
{ "id": "ch01-christy-signing",
  "file": "ch01-christy-signing.jpg",
  "page": 9,
  "caption": "\"Scene at the Signing of the Constitution,\" by Howard Chandler Christy.",
  "credit": "Courtesy of the Library of Congress.",
  "chapters": ["constitution"] }
```

Curation drops page furniture, logos, gradients and color-separation artifacts
(the `index`-color entries in `pdfimages -list` are usually these), and **drops
the AP-credited image**. Every surviving image renders with caption and credit
visible. Downscale to ~1600px max and re-encode to keep the binary reasonable.
-> *verify:* every manifest `file` exists in `embed.FS`; every embedded image
appears in the manifest (no orphans); every entry has a non-empty `credit`; no
entry credits the Associated Press; every `chapters` ref resolves.

**2.5 Officials + states (bundled snapshots)** — `officials.json` (`as_of` +
President, VP, Speaker, Chief Justice, all 50 governors) and `states.json`
(50 states + DC + territories -> capital, Wikidata QID, senators, plus the
DC/territory special-case wording the PDF spells out verbatim). `zcta_cd.txt.gz`
is produced by `cmd/ingest` from the Census relationship file.
-> *verify:* every `CurrentOfficial` question resolves from `officials.json`;
every `StateSpecific` question resolves for all 50 states + DC; DC and territory
edge cases reproduce the PDF's exact wording (D.C. has no senators, no governor,
is not a state); the crosswalk contains a known-good sample of ZIPs including at
least one four-district ZIP.

**2.6 `required_counts.json`** — the authored cardinality table. Only the seven
enumeration questions deviate from 1:

```json
{ "10": 2, "48": 2, "65": 3, "67": 2, "69": 2, "81": 5, "126": 3 }
```

-> *verify:* (a) every listed question's `RequiredCount` <= its number of
answers; (b) **a guard test that greps every prompt for a cardinality word
("name two", "what are three", "name five") and asserts each hit is either in
this table or on an explicit reviewed-exceptions list** — so if USCIS revises
the PDF and adds an enumeration question, the test fails loudly instead of the
app silently accepting one answer where three are required. The reviewed
exceptions are the wording-lies cases: Q15, Q16, Q19, Q28, Q37 and friends.

---

## Phase 3 — Engines

**3.1 Grading** (`internal/quiz/grade.go`). The real test is **oral**, judged by
an officer, so grading is advisory and generous, never punitive.

Implemented: `Grade` recognizes fixed answers, returns the accepted items and
how many distinct items are still needed, and marks changing-answer prompts as
unavailable until officials resolution exists. Its hermetic tests require every
fixed official bullet to self-match as an item, enforce every enumeration's
cardinality and distinctness, and cover optional text, digit/word equivalence,
U.S. spelling variants, diacritics, guidance exclusion, the judicial/judiciary
term variant, and a bounded one-typo or transposition tolerance for substantial
words. Semantic rules are deliberately explicit and source-reviewed rather
than model-guessed: the first recognizes the Study Guide's stated roles of the
three branches (making, enforcing, and reviewing federal laws) for Q16. A
separate constrained rule applies only to single-answer prompts: it can accept
an oral response with one omitted meaningful term when at least two remain
(for example, Q110/Q111 “stop communism” for “stop the spread of communism”);
multi-answer menus retain strict distinct-item matching. The reviewed rule
catalog also covers the Study Guide's alternate wording for Q3, Q8, Q18, Q20,
Q47, Q84, Q95, Q106, and Q112; every rule carries its source passage and a
test that pins its target accepted answer.

*Normalization* (applied to both user input and every acceptable answer):
lowercase; strip punctuation and diacritics; collapse whitespace; drop leading
articles; normalize digits <-> words ("27" == "twenty-seven", "2" == "two");
normalize "U.S." / "US" / "United States".

*Single-answer match* (`RequiredCount == 1`) — input matches if, for any
`Answer`, it equals `Core` or `Full`, or contains `Core` at token boundaries.
Containment handles "the U.S. Constitution" against `(U.S.) Constitution`.

*Set match* (`RequiredCount == n`) — this is where the app earns its keep:

1. Split the user's input on commas, semicolons, newlines and " and ".
2. Match each fragment against the answer list independently.
3. **Enforce distinctness** — two fragments matching the *same* `Answer` count
   once. Answering "New York, New York, New York" to Q81 is one state, not five.
4. Report `matched k of n required`, naming which items landed and how many are
   still needed, rather than a bare wrong.
5. Partial credit is shown but does not pass the question.

*Override* — an explicit **"I got this right"** button on every question. The
user is the authority; the override feeds progress tracking. This is the safety
valve for any normalization gap.

-> *verify:*
- **Self-match test:** for all 128 questions, every official answer string must
  grade as correct when submitted verbatim. This is the single most important
  test in the codebase.
- **Enumeration tests:** for each of the seven, submitting exactly N distinct
  valid items passes; N-1 fails with the right "need 1 more" message; N copies
  of the same item fails on distinctness; N valid items in any order passes;
  mixed separators ("Vote and run for office") parse correctly.
- **Optional-paren tests:** `Constitution` and `U.S. Constitution` both pass Q2.
- **Guidance tests:** bracketed reader instructions are never treated as
  acceptable answers.
- **Near-miss table:** realistic misspellings and paraphrases, each with an
  asserted expected verdict, so grading changes are visible in the diff.

**3.2 Quiz engine** (`internal/quiz/engine.go`):

Implemented: `NewSession` draws deterministic, no-repeat official and 65/20
sets from an injected RNG; `Record` stops a session when passing or failing is
mathematically certain. The web question flow is implemented; custom drills
remain.
- **Official mock:** 20 drawn from all 128, pass at 12. Stops early once pass or
  fail is mathematically determined, mirroring the real interview.
- **65/20 mode:** 10 drawn from the 20 starred, pass at 6.
- **Custom drill:** filter by section, chapter, or "questions I've missed."
- **Multiple-choice mode:** only offered for `RequiredCount == 1` questions —
  distractors sampled from *other* questions' answers in the same subsection.
  Enumeration questions stay free-text, since a multiple-choice "pick three"
  would misrepresent the real oral format.
-> *verify:* seeded-RNG determinism; 20 unique questions, never a duplicate;
65/20 draws only from the starred set; threshold arithmetic at boundaries
(11 vs 12, 5 vs 6); early-stop fires at the right moment; distractors never
equal a correct answer; MC mode never selects an enumeration question.

**3.3 Flashcard scheduler** (`internal/flashcard/scheduler.go`) — Leitner boxes
(5 boxes, intervals 1/2/4/8/16 days). Correct -> promote; wrong -> box 1. Clock
injected as an interface so tests are deterministic. Enumeration cards show the
full menu on the reverse with the required count stated ("name any 3 of these 11").
Implemented with an all-questions, 65/20, and missed-more-often deck picker.
-> *verify:* promotion/demotion transitions; due-date computation across day
boundaries; a card answered correctly 5 times stops appearing daily; deck filters
(section, chapter, 65/20, starred, missed) return the right sets.

**3.4 Store** (`internal/store`) — attempt history, per-card box state, starred
questions, state profile.
Implemented as a local JSON file with atomic temp-file replacement, recovery
from missing/corrupt data, and mutex-protected updates. Flashcard progress is
stored at the platform config location (`n400/progress.json`); no street
address is ever stored.
-> *verify:* round-trip; atomic write leaves no partial file on simulated
failure; corrupt or missing file recovers to empty state rather than crashing;
concurrent access clean under `-race`.

**3.5 Officials resolution** (`internal/officials`) — the overlay chain
(manual override -> live fetch -> bundled snapshot), staleness computation, and
ZIP -> district lookup over the embedded crosswalk.
-> *verify:* overlay precedence in every combination; staleness boundary at
exactly the threshold day; a four-district ZIP returns four candidates; an
unknown ZIP returns a clean "not found" rather than an error; DC/territory ZIPs
produce the special-case wording.
**All network clients are tested against `httptest` servers replaying recorded
fixtures — no test ever touches the live network.** Fixture tests cover: happy
path, HTTP 500, timeout, malformed JSON, empty result set, and a schema change
(unexpected field shape), each asserting the previous value survives and the
failure is reported rather than swallowed.

Implemented so far: the resolver honors manual -> sidecar -> bundled precedence,
labeling the sidecar's resolved answers with its fetch date; the dated snapshot
covers four federal offices, capitals, and state senators; `states.json` now
also carries each state's Wikidata QID (D.C. deliberately has none); and the
embedded 119th Congress Census crosswalk returns all ZIP candidates. A
Wikidata `FederalClient` (all four federal offices) and `GovernorClient` (one
state's governor, by QID) each perform a bounded query through an injected
HTTP client, with happy-path and HTTP-failure fixture tests; both send the
`User-Agent` Wikidata's usage policy requires and decode only the fields they
need, since a real response nests additional standard SPARQL fields a first
strict-decoding attempt had rejected outright — caught only by testing against
the live endpoint, not by the fixtures, which is now corrected in both the
code and the fixtures. The Settings **Refresh current officials** action runs
both clients — federal always, the governor only for the learner's currently
selected state, skipped without error when that state has no QID (D.C.) — and
atomically writes `officials-live.json` as the sidecar, preserving any other
state's previously fetched governor; **Reset to bundled** removes it entirely;
all of this is route- and fixture-tested, including a failed refresh (from
either client) leaving a prior sidecar value intact and rendering its error
inline. `--offline` disables the refresh route only — reading an existing
sidecar is local disk I/O, not network, so it stays available offline.

The House roster (`internal/officials/data/house.json`) is bundled the same
way senators are: fetched once from `unitedstates/congress-legislators`,
dated, and checked in rather than live-refreshed, since representatives don't
change on a schedule frequent enough to justify the added client and the
resolver already supports a manual override for the day a seat turns over
early. `Snapshot.Representative(stateCode, district)` resolves Q29: an exact
`"ST-DD"` match first, falling back to a state's sole roster entry when it has
only one (every at-large state, and D.C.'s non-voting delegate) — needed
because the ZIP crosswalk's Census-assigned district number for those seats
doesn't always match congress-legislators' own numbering (D.C.'s crosswalk
entries use district "98"; congress-legislators lists its delegate as
district "0"). `Resolver` gained a `District` field, set from the stored
profile the same way `Manual`/`Sidecar` already are, so `ResolveAll`'s
signature didn't need to change.

`GeocoderClient` (`fetch_geocoder.go`) resolves a street address to a
district via the Census Bureau's keyless geocoder, for a learner who opts
into it to disambiguate a multi-district ZIP instead of picking from the
crosswalk's candidate list; the address is sent to Census but never stored,
only the resulting district. It requires exactly one address match and
exactly one congressional-district geography layer in the response, erring
rather than guessing when either is ambiguous or absent. Like the ZIP
crosswalk, it reads the district from the layer's `GEOID` field (state FIPS +
district) rather than the Congress-numbered `CD119`/`CD120`-style field,
since `GEOID`'s name is stable across a Congress-number rollover while the
layer's own name and per-vintage field name are not — confirmed live: the
Census endpoint had already moved on to serving `"120th Congressional
Districts"` by the time this was built, not `"119th"`. Wired into Settings as
`POST /settings/address`, shown only once a ZIP has already resolved to more
than one district candidate, gated by `--offline` the same way refresh is,
and rendering its error inline on failure without disturbing the ZIP/district
already on file. Still required before release: timeout/malformed/empty/
schema-drift fixtures for all three network clients (Wikidata's two and the
geocoder).

---

## Phase 4 — Web UI

Hand-written design system in `static/app.css`: CSS custom properties for the
palette, light **and** dark via `prefers-color-scheme`, a type scale, responsive
down to ~380px since "cross-platform" includes reading on a phone. Restrained
palette — deep navy, off-white, one red accent — not flag cosplay. System font
stack plus one self-hosted serif for chapter prose.

| Route | Purpose |
|---|---|
| `GET /welcome` · `POST /welcome` | First-run wizard: state + ZIP -> district (with candidate picker if the ZIP spans several), optional address disambiguation, optional initial refresh. Skippable. |
| `GET /` | Dashboard: progress, due flashcards, next chapter, "take a test" |
| `GET /learn` · `/learn/{chapter}` | Chapter list; reader with objectives, prose, images+credits, linked questions |
| `GET /library` · `/library/{doc}` | Reference browser: document list; reader with prose, linked questions. Deep links into a specific Article/Section/Amendment use in-page anchors on the Constitution document, the same way chapter sections do, rather than a separate nested route. **Implemented** for the Declaration, Constitution, Amendments, and the first Almanac document; the rest of the Almanac remains. |
| `GET /flashcards` · `POST /flashcards/{id}/answer` | Deck picker; card flip (Alpine) + HTMX answer posting |
| `GET /practice` · `POST /practice/start` · `/practice/{session}` | Implemented local-only mode picker, free-text question flow, advisory grading, self-check override, early pass/fail result, and official/65-20 modes. Attempts are intentionally not persisted yet; review links and custom drills remain. |
| `GET /questions` · `/questions/{id}` | Browse all 128; single question with all acceptable answers, required count, chapter links, ⚠ banner if changeable |
| `GET /settings` · `POST /settings/refresh` | State/ZIP/district profile, manual answer overrides, **Refresh current officials**, reset to bundled, content `as_of` dates |

UI must make cardinality obvious wherever a question appears: an enumeration
question renders **"Name 3 — any three of the following"** above its input, and
the input shows live progress ("2 of 3 named") as the user types.

-> *verify:* `httptest` over every route (status, content-type, key content);
golden HTML snapshots for the chapter reader and results page; 404s for unknown
chapter/question/session IDs; a test asserting every changeable-answer question
renders the uscis.gov banner; a test asserting all seven enumeration questions
render their required count.

---

## Phase 5 — Build and release

- `Makefile`: `test` (`go test -race ./...`), `lint` (`go vet` + staticcheck),
  `build`, `build-all` (linux/darwin/windows × amd64/arm64), `ingest`.
- README: how to run, how to re-ingest when USCIS updates the PDFs, how to
  refresh `officials.json`.
- Test asserting the binary is self-contained: server starts and serves `/` with
  the working directory set to an empty temp dir.

---

## Critical files

| File | Role |
|---|---|
| `internal/content/parse_questions.go` | Ground truth. Everything depends on it being exactly right. |
| `internal/quiz/grade.go` | Subtlest logic in the app. Wrong here = the app tells you you're wrong when you're right. |
| `internal/content/data/required_counts.json` | Seven questions where "any one answer" is not enough. Authored, guard-tested. |
| `internal/content/validate.go` | Cross-reference integrity net (questions <-> chapters <-> images <-> library). Runs as a test **and** at server startup. |
| `internal/content/data/chapters/*.md` | Bulk of the authoring work; carries the question->chapter mapping. |
| `internal/content/data/officials.json` | The bundled snapshot — the thing that goes stale. Dated, banner-flagged, refreshable. |
| `internal/officials/resolve.go` | The overlay chain (override -> sidecar -> live -> bundled), expiry computation, and provenance labelling for all eight changing answers. |

## Verification

```bash
go test -race ./...     # unit + golden + httptest; hermetic (no PDFs, no poppler)
go vet ./...
go build ./cmd/n400 && ./n400
```

Manual smoke before calling it done:

1. Read a chapter — prose intact, images render with credits, linked questions appear.
2. Run a full 20-question mock — threshold correct, results link back into Learn.
3. Run 65/20 mode — only starred questions appear.
4. **Answer Q81 with "New Hampshire, Massachusetts, Rhode Island, Connecticut,
   New York"** — passes. Answer it with the same state five times — fails on
   distinctness with a clear message.
5. **Answer Q2 with "Constitution"** and again with "the U.S. Constitution" —
   both pass.
6. Flashcard a deck, close the app, reopen — progress survived.
7. Run the first-run wizard with a Texas ZIP — correct district and
   representative. Re-run with **90002** (four districts) — the candidate picker
   appears and does not guess. Use its **Look up my district** address form
   instead — a full street address in that ZIP resolves the exact district
   and representative without picking from the list.
8. Set state to DC — Q23/Q61/Q62 show the PDF's exact special-case wording.
9. Hit **Refresh current officials** — President/VP/Speaker/Chief Justice/governor
   update and show source + fetch date. Unplug the network and hit it again —
   clear per-source failure message, previous values intact, app still usable.
10. Confirm the uscis.gov banner is present on all eight changing questions in
   every state: bundled, freshly fetched, and hand-overridden.
11. Run with `--offline` — no outbound connections (verify with `ss`/`tcpdump`),
   everything else works.
12. Open `/learn` at phone width — no horizontal scroll.
13. Disconnect the network entirely — every section still works.

## Risks

- **Chapter transcription is the long pole.** 12 chapters of two-column PDF prose
  needing manual reflow. Mitigation: do Chapter 1 end-to-end first to calibrate
  effort before committing to all 12.
- **Grading generosity is a judgment call.** Too strict and the app punishes
  correct answers; too loose and it builds false confidence. Mitigation: the
  self-match test as a hard floor, the near-miss table as a visible record of
  where the line sits, and the "I got this right" override as the escape hatch.
- **Question->chapter mapping is authored, not derived.** Validation catches
  *omissions*, not *bad* mappings — worth a read-through once complete.
- **Upstream sources can change shape or go away.** Wikidata is
  community-edited; the Census and legislators files are versioned by Congress
  number. Mitigated by: the bundled snapshot always being a working fallback,
  the `N-1` ingest fallback for an unpublished Census file, fixture tests that
  fail loudly on schema drift, refresh failures being non-fatal by construction,
  and the sidecar file as a no-infrastructure manual override. The app degrades
  to "offline with a stale-data banner", never to broken.
- **Nobody presses the refresh button.** The likeliest real failure is a user
  studying against January-2027 data in March 2027. Mitigated by deriving a hard
  expiry date from the data instead of guessing, warning *before* it lands, and
  flagging affected answers individually rather than relying on one banner the
  user dismissed months ago.
- **Wikidata is not USCIS.** A community edit could briefly show a wrong
  officeholder. Mitigated by the permanent verification banner, visible source +
  fetch-date labels, and the manual override. This is why refresh is
  user-initiated and never silent.
- **Privacy.** ZIP and state are stored locally only. A street address, if the
  user opts into that path, goes to the Census Bureau and is never persisted —
  stated plainly in the wizard, not buried.
- **Image curation is manual.** 110 extracted files, most of them page furniture.
  Budget real time for a contact sheet and picking keepers.
