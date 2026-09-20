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

Chapter 1 was reviewed visually against supplied PDF pages 8–17. Page 17 is blank
except the footer. Source word coverage is checked per page against checked-in
raw text, with a full parsed golden in testdata. This is not a substitute for
visual source review: counts alone cannot validate reading order or diagram edges.

Only the three credited images are reproduced. The other photographs and header
art have no printed institutional credit and remain omitted. Diagram text is
retained, including the printed typo Respresentatives. The page 16 map labels
are transcribed, including its two Massachusetts labels and Atlantic Ocean;
the map's geography is not reproduced. This is a text edition, not a facsimile.

Question links are authored from topics actually covered; Chapter 1 covers 25
questions. Questions without a published chapter do not get speculative links.
The all-128 chapter-coverage requirement is a release gate for the full 12-chapter
corpus, not satisfied by this first chapter.
