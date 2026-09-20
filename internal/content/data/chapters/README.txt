Chapter authoring format

Metadata between --- lines is JSON (a YAML subset), decoded strictly by the
standard library. Unknown fields are errors. Body markup deliberately supports
only ## and ### headings, one-line reflowed paragraphs, - lists with one nested
level, ![printed caption](manifest-id), fenced text diagrams, :::sidebar blocks,
and > printed captions for illustrations not reproduced. No raw HTML is trusted.

Each source page begins with <!-- page:N -->. Source page markers are increasing
and lie within source_start/source_end. Figure captions and credit lines live in
the image manifest; the Markdown caption must match exactly. Source-only notes
about transcription choices belong in editorial_notes, not in the book's prose.

Each published chapter is reviewed visually against its supplied PDF pages
(Chapter 1: 8–17; Chapter 2: 18–23; Chapter 3: 24–28; Chapter 4: 29–32). Source
word coverage is checked per page against checked-in raw text, with a full
parsed golden in testdata. This is not a substitute for visual source review:
counts alone cannot validate reading order or diagram edges. A diagram with no
extractable PDF text layer (Chapter 2's page 22 lawmaking flowchart) has its
transcribed labels recorded separately in
`data/raw/study-guide-visual-supplement.json` and checked by the same coverage
test.

The coverage test strips page furniture (page numbers, the running footer, the
"CHAPTER N[: TITLE]" banner for any chapter, the "In this chapter..." line, and
the chapter's own title line on its opening page only) before comparing word
inventories. A chapter title is only stripped on its opening page: a later page
may legitimately reuse the exact title text as a genuine section heading
(Chapter 4's page 31), and that occurrence must still be counted.

Only images with a printed institutional credit are reproduced (five so far,
all in Chapters 1–3; Chapter 4 has none). The other photographs and header art
have no printed credit and remain omitted, with any printed caption retained as
reader text. Diagram text is retained, including printed typos and
inconsistencies (e.g. Chapter 1's Respresentatives, Chapter 3's uneven
icon-caption punctuation). Chapter 1's page 16 map labels are transcribed,
including its two Massachusetts labels and Atlantic Ocean; the map's geography
is not reproduced. These are text editions, not facsimiles.

Question links are authored from topics actually covered, and a question whose
content is genuinely repeated across chapters (e.g. Q18, Q41, Q50) links to all
of them. Questions without a published chapter do not get speculative links.
Chapters 1–4 cover 25, 20, 12, and 8 questions respectively. The all-128
chapter-coverage requirement is a release gate for the full 12-chapter corpus,
not satisfied yet.
