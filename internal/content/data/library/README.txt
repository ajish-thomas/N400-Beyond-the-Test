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
page only, a line matching the document's own title in full caps (the
printed running head), the same treatment chapters give their opening title
splash.

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

Remaining: the Constitution's Preamble, Articles I-VII (with Sections), and
Amendments I-XXVII, all from the same PDF; footnoted superseded clauses
(e.g. Article I Section 3's original method of electing senators, changed by
the Seventeenth Amendment) and each amendment's ratification date are
printed as page-bottom footnotes in the source and will be inlined as plain
text at their point of reference rather than a separate footnote UI, since
this is a text edition, not a facsimile. Then the Citizen's Almanac's seven
speeches, symbols/anthems, and four landmark Supreme Court cases, each
carrying the required citation.
