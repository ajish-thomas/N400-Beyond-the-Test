# CLAUDE.md

Guidance for Claude Code (claude.ai/code) when working in this repository.

## Project Overview

**Goal:** A study application for the USCIS civics test (part of the N-400
naturalization application) that does two jobs at once:

1. **Pass the test.** Drill the official 128 questions until answers are automatic.
2. **Actually learn.** Go past flashcard level into a real guide to American
   government and history, so the material means something.

**Current state:** Phase 1 plus the first Phase 2 chapter: extraction pipeline,
validated question corpus, offline question browser, selectable light/dark theme,
and Chapter 1 text reader with 25 question links and three credited images.
See `README.md` for implemented scope,
verification commands, source discrepancies, and remaining phases.
`PLAN.md` is the detailed specification — read it before writing code.

## Technology decisions (settled — do not relitigate)

| Area | Decision |
|---|---|
| Language | **Go**, single self-contained binary, no cgo |
| Frontend | Go `html/template` + vendored HTMX/Alpine + hand-written CSS |
| Build | **`go build` is the entire build.** No Node, no npm, no bundler |
| Assets | Everything via `embed.FS` — content, images, CSS, JS |
| Runtime | Local HTTP server on `127.0.0.1:8400`, opens default browser |
| Persistence | JSON file in `os.UserConfigDir()/n400/`, atomic writes |
| Network | **Offline-first.** Fully usable unplugged; network is enhancement only |
| Platforms | linux/darwin/windows × amd64/arm64 |

## Application sections

- **Learn** — 12 chapters transcribed from the Study Guide, each with learning
  objectives, images with credits, and the test questions it covers.
- **Library** — Declaration, Constitution (Articles I–VII), all 27 Amendments,
  7 historical speeches, symbols/anthems, 4 landmark Supreme Court cases.
- **Flashcards** — 128 cards, Leitner scheduling, filterable decks.
- **Practice Test** — official-format mock (20 questions, pass at 12) and
  65/20 mode (10 starred questions, pass at 6).

## Source materials (ground truth)

| File | Role |
|---|---|
| `2025-Civics-Test-128-Questions-and-Answers.pdf` | **Ground truth** for questions and answers |
| `USCIS-2025-Civics-Test-Study-Guide.pdf` | Chapter prose + images (88pp, 40MB) |
| `CitizensAlmanac-M-76.pdf` | Speeches, symbols, landmark cases (**text only** — see AGENTS.md) |
| `DOI-Constitution-M-654.pdf` | Declaration + Constitution full text |

## Planned layout

```
cmd/n400/          # the app
cmd/ingest/        # dev-only PDF extraction (needs poppler)
internal/content/  # parsers, models, validation, embedded data/
internal/quiz/     # selection, scoring, answer grading
internal/flashcard/# deck building, Leitner scheduler
internal/officials/# the 8 changing answers: lookup, overlay, expiry
internal/store/    # progress persistence
internal/web/      # server, handlers, templates, static
```

## Two things that make this project unusual

**1. Answers are sets, not strings.** Each question has a list of *acceptable*
answers plus a required count. Most need any one; seven need 2, 3, or 5 distinct
items. A number in the question wording does **not** imply the required count.
This is the single easiest thing to get wrong — see `PLAN.md` § "The answer model".

**2. Eight answers have a shelf life.** President, VP, Speaker, Chief Justice,
your senators, your representative, your governor, your state capital. These are
bundled as dated offline snapshots with a user-initiated refresh, a derived hard
expiry date, and a permanent "verify at uscis.gov" banner. See `PLAN.md`
§ "Answers that change" and § "Staying current".

The other 120 questions, all 12 chapters, and the whole library never go stale.

## Development philosophy

1. **The PDFs are the source of truth.** The app extracts and presents their
   content; it never replaces, paraphrases, or supplements it with invented facts.
2. **Correctness before scope.** A study app that marks a right answer wrong is
   worse than one missing a feature.
3. **Simplicity over features.** Build what's needed to study effectively.
4. **Offline is the default, not a fallback.**
5. **Test everything.** Content accuracy is testable — so test it.

See `AGENTS.md` for implementation standards, content rules, and testing requirements.
