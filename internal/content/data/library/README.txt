Library document authoring format

Library documents (the Declaration of Independence, the Constitution and its
amendments, and — from the Citizen's Almanac — speeches, symbols/anthems, and
landmark Supreme Court cases) share the exact Markdown grammar and JSON front
matter used by Learn chapters (see `data/chapters/README.txt`): the same
`##`/`###` headings, one-line reflowed paragraphs, `- ` lists, `<!-- page:N
-->` markers, and `> ` captions, parsed by the same `splitDocument`/
`parseBlocks` functions in `chapters.go`. `ParseLibraryDoc` (`library.go`)
decodes front matter into `LibraryDoc` instead of `Chapter`; it has no
chapter number, objectives, or images, and adds three fields chapters don't
need:

- `kind`: `founding-document`, `speech`, `symbol`, or `case`.
- `source`: `declaration-constitution` (from `DOI-Constitution-M-654.pdf`) or
  `citizens-almanac` (from `CitizensAlmanac-M-76.pdf`).
- `citation`: required, and checked against the exact string AGENTS.md
  specifies, when `source` is `citizens-almanac`; must be empty otherwise.
  Almanac text is reusable only with this citation attached; a document that
  gets it wrong or omits it fails to parse rather than publish uncredited.

Images are not supported in library documents (`catalog.go`'s validate()
rejects any `image` block here): the Declaration/Constitution booklet's two
images are cover art, not captioned illustrations, and Almanac images are
never usable at all (its own notice states they are not public domain
outside that publication) — this project's `--images`/`--curated-images`
pipeline in `cmd/ingest` never extracts from it.

Source word coverage is checked per page against checked-in raw text
(`data/raw/constitution.txt` or `data/raw/almanac.txt`, chosen by `source`),
the same technique `TestChapterSourceCoverage` uses for the Study Guide,
implemented as `TestLibrarySourceCoverage`. Page furniture here is simpler
than the Study Guide's: bare page-number lines and, on a document's opening
page only, one or more lines whose concatenation matches the document's own
title in full caps (the printed running head — a single line for the
Declaration, two for "THE CONSTITUTION" / "OF THE UNITED STATES OF AMERICA"),
the same generalized multi-line handling chapters give a title splash that
spans more than one line. `us-constitution.md` needs one more furniture
rule: the source's own footnote reference digits (see below) are stripped
from the raw side via a small, exact-match `footnoteRefs` map keyed by page,
the same targeted-substitution technique chapters' `mapLabelFixes` uses for
PDF-extraction artifacts — without it, a bracket-attached digit like
"...Persons.]1" tokenizes as the word "1", which the authored side no
longer has once the reference is inlined as a parenthetical.

The Declaration of Independence (`declaration.md`) is the first document.
Its source pages (`DOI-Constitution-M-654.pdf`, pages 7-13) are single-column
and did not need the two-column reflow Study Guide chapters require, but
three sentences still split across a page boundary in the raw extraction
(the "WE hold these Truths..." paragraph across pages 7/8, "He is, at this
Time, transporting large Armies..." across pages 10/11, and the closing
"...We must, therefore, acquiesce..." across pages 11/12); each is authored
as a matching split across two blocks, the same treatment chapters give a
mid-sentence page break. The signature block (page 12) is transcribed as
four short lines rather than run into prose, and the signers list (page 13)
preserves the source's own state-by-state grouping and order (Georgia
through Connecticut) as thirteen `###` subheadings with a list each, rather
than one flattened list. Original 1776 spelling and capitalization are kept
exactly as printed (e.g. "shewn", "compleat", "harrass"), not modernized.

Question links are authored from what a document's own text literally
states, the same discipline chapters use. The Declaration links five
questions (Q8, Q9, Q10, Q11, Q79) whose accepted answers it states directly.
Q78 ("Who wrote the Declaration of Independence?") is deliberately not
linked: the document names Thomas Jefferson only as a Virginia signer, never
stating that he drafted it.

The Constitution's Preamble and Articles I-VII (`us-constitution.md`,
pages 15-38 of the same PDF) are the second document. It also includes,
for completeness, the Constitution's own signers list (page 33 — 12 states,
since Rhode Island sent no delegates to the Convention) and two further
pages the source prints immediately after the Articles and before the
Amendments: the ratifying convention's closing resolution ("In Convention
Monday", pages 35-36) and the first Congress's resolution transmitting the
proposed Bill of Rights (pages 37-38, under the heading "Congress OF THE
United States"). Eleven sentences split across a page boundary in the raw
extraction (more than the Declaration's three, since these pages run denser
and longer); each is authored as a matching split block, the same treatment
as the Declaration. The source prints 11 numbered footnotes within these
pages, each marking text later changed or superseded by a specific
amendment (e.g. the original method of electing Senators, changed by the
Seventeenth Amendment) or, for footnote 5, pointing forward to the Sixteenth
Amendment; each is inlined as a parenthetical immediately after the bracket
or word it annotates (see `footnoteRefs` above) instead of a separate
footnote/superscript UI. Footnote 11 is the one exception: printed on page
37 attached to a heading rather than a body clause, and several sentences
long, it is transcribed as its own paragraph immediately after that heading.
Two printed inconsistencies are kept exactly as found: "Article III." (page
28) is missing the period after "Article" every other Article heading has,
and George Washington's signature is abbreviated two different ways in two
different places. This document links 10 questions (Q2, Q5, Q16, Q17, Q19,
Q20, Q25, Q36, Q42, Q50); several topically close questions (Cabinet, veto,
"for life") are deliberately not linked because the 1787 text never uses
those specific words, and none is linked for a fact whose current legal
authority is one of the superseded, bracketed clauses (e.g. presidential
succession) pending the still-unauthored Amendments document.

Remaining: the Constitution's Amendments I-XXVII, from the same PDF —
their own ratification-date footnotes will need the same inlining
treatment as `us-constitution.md`'s — then the Citizen's Almanac's seven
speeches, symbols/anthems, and four landmark Supreme Court cases, each
carrying the required citation.
