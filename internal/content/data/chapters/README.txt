Chapter authoring format

Metadata between --- lines is JSON (a YAML subset), decoded strictly by the
standard library. Unknown fields are errors. Body markup deliberately supports
only ## and ### headings, one-line reflowed paragraphs, - lists with one nested
level, ![printed caption](manifest-id), fenced text diagrams, fenced verse
blocks (```verse; library documents only so far, for the Citizen's Almanac's
songs and poems — see data/library/README.txt), :::sidebar blocks, and >
printed captions for illustrations not reproduced. No raw HTML is trusted.

Each source page begins with <!-- page:N -->. Source page markers are increasing
and lie within source_start/source_end. Figure captions and credit lines live in
the image manifest; the Markdown caption must match exactly. Source-only notes
about transcription choices belong in editorial_notes, not in the book's prose.

Each published chapter is reviewed visually against its supplied PDF pages
(Chapter 1: 8–17; Chapter 2: 18–23; Chapter 3: 24–28; Chapter 4: 29–32;
Chapter 5: 33–37; Chapter 6: 38–42; Chapter 7: 43–46; Chapter 8: 47–51;
Chapter 9: 52–57; Chapter 10: 58–61; Chapter 11: 62–68; Chapter 12: 69–76).
Source word coverage is checked per page against checked-in raw text, with a
full parsed golden in testdata.
This is not a substitute for visual source
review: counts alone cannot validate reading order or diagram edges, though
it does catch dropped/added words immediately — an early Chapter 5 draft
that dropped "in federal elections" from one sentence failed this test right
away. A diagram or map label with no *legible* extractable PDF text layer
(Chapter 2's page 22 lawmaking flowchart; Chapter 6's page 39 "Washington,
D.C." marker, page 40 "GULF OF AMERICA," and page 41 mountain/river names;
Chapter 7's page 45 "Atlantic Ocean" label; Chapter 11's page 65 Cold War map
"Atlantic Ocean" label, all rendered as scrambled fragments by curved or
diagonal source text) has its clean transcription recorded separately in
`data/raw/study-guide-visual-supplement.json` and checked by the same
coverage test. Normalizing a printed credit to its allowlisted phrase can
also introduce words absent from the source (Chapter 11's page 67 Pentagon
flag photo, printed as "White House photo by Paul Morse" with no "Courtesy
of..." phrase at all); the invented words ("Courtesy", "of", "the") are
likewise recorded in the visual supplement rather than silently unaccounted
for. A sentence that itself splits across a
page boundary in the source (Chapter 5's Federalist Papers sentence, pages
33/34; Chapter 8's causes-of-the-war sentence, pages 47/48) is authored as a
matching split across two blocks, not smoothed into one. A page whose PDF
text layer contains a hidden duplicate of a visible label (Chapter 6's page
41 duplicates "Alaska," "Hawaii," and all five territory names in ALL CAPS
behind the real mixed-case labels — confirmed against the rendered page to
have only one instance of each printed) has the coverage test drop the extra
duplicate rather than this edition transcribing a label that isn't actually
there.

The coverage test strips page furniture (page numbers, the running footer, the
"CHAPTER N[: TITLE]" banner for any chapter and any title separator — Chapter
6 uses an en dash where Chapters 1–5 used a colon — the "In this chapter..."
line, and the chapter's own title splash on its opening page only) before
comparing word inventories. The title splash is only stripped on its opening
page: a later page may legitimately reuse the exact title text as a genuine
section heading (Chapter 4's page 31), and that occurrence must still be
counted. The splash usually prints as one line, but Chapter 8's prints across
two ("THE AMERICAN REVOLUTIONARY WAR &" / "THE DECLARATION OF INDEPENDENCE");
the filter scans the whole span between the "CHAPTER N" banner and "In this
chapter..." and matches its full concatenation against the title, rather than
requiring one single matching line, so both one-line and multi-line splashes
are handled the same way.

Chapter 12's last source page (76) is genuinely blank in the PDF beyond the
running footer, which the coverage test already strips; its front matter's
`source_end` covers the page anyway, and the last `<!-- page:N -->` marker
is simply the previous page (75), which the parser allows since it only
requires markers to be increasing and within range, not exhaustive.

Only images with a printed institutional credit that has been reviewed for
licensing are reproduced from the Study Guide itself (thirty-one so far, in
Chapters 1–3 and 7–12; Chapters 4–6 have none from the PDF). The
reviewed-federal-institution allowlist in `catalog.go` grew by one for
Chapter 7 ("Courtesy of the National Park Service.") and again for Chapter 8
("Courtesy of the U.S. Senate.", reviewed as a federal legislative body's own
art collection), then twice more for Chapter 11 ("Courtesy of NASA." for the
Apollo 11 Moon-landing photo; "Courtesy of the White House." for the
Pentagon flag photo, the Executive Office of the President) — the first
additions since Chapter 2's JFK Presidential Library and Museum entry.
Chapter 12 reuses the U.S. Senate and Library of Congress credits without
adding a new allowlist entry. A printed credit alone is not enough: Chapter 5's
voting-booth photo and one of Chapter 7's four photographs are each credited
to a non-federal source ("the Polling Place Photo Project"; "the Jamestown
Yorktown Foundation," a Virginia state institution) that has not been
reviewed, so each is omitted like an uncredited image, with its full caption
(credit line included) kept as reader text pending that review. Other
photographs and header art with no printed credit at all are handled the
same way — Chapter 8's Benjamin Franklin portrait is the only one of its
seven photographs/paintings omitted for this reason (the other six are all
reproduced), Chapter 9 omits five of its six for the same reason (keeping
only its Constitution document, already allowlisted), and Chapter 10 omits
four of its eight (keeping cotton picking, Susan B. Anthony, Frederick
Douglass, and the Nast "Emancipation" print — the latter's own credit line
is missing the trailing period Chapter 10's other Library of Congress
credits use; the manifest field holds the standard normalized string, and
the punctuation gap is noted rather than silently carried through), and
Chapter 11 omits five of its fifteen (the Cold War map, Dr. Martin Luther
King Jr., Dolores Huerta, the March on Washington, and the September 11
rescue-workers photo), and Chapter 12 omits eleven of its fifteen the same
way. Chapter 12 additionally has one captioned illustration that carries a
printed credit but is still not reproduced: "Raising the Flag on Iwo Jima"
is credited to "the Associated Press," which `catalog.go`'s validate()
function has rejected by name since this project's first chapter — this is
the first chapter where that standing rule actually excludes an image
rather than sitting unused. Two more graphics on Chapter 12's pages (the
opening banner and the full flag illustration on page 71) have no caption
at all and aren't counted as illustrations, the same treatment given every
chapter's uncaptioned opening banner since Chapter 1. Unlike every image
chapter since Chapter 1, none of Chapters 9 or 10's map/photo graphics that
lack a credit have any extractable text layer beyond their captions — no
state names or other labels — so no visual-supplement entry was needed for
either chapter; Chapter 11's uncredited Cold War map is the exception,
since its "United States" and "Soviet Union" labels are cleanly extractable
and transcribed as a diagram even though the map's caption itself is
omitted for lack of a credit. A credit can also carry an item number
beyond the standard allowlisted phrase (Chapter 8's Currier & Ives print:
"...Courtesy of the National Archives, 532915."); the number is moved into
the caption so the credit field holds the exact allowlisted string, losing
no text. Diagram text is retained, including printed typos and
inconsistencies (e.g. Chapter 1's Respresentatives, Chapter 3's uneven
icon-caption punctuation, Chapter 5's "Civils Rights Act of 1964" and a
dropped "and" in "raise their right hand say the Oath of Allegiance,"
Chapter 6's "GULF OF AMERICA," Chapter 7's extra "the" in "the North
America", Chapter 9's tribes list including "Arawak" and "Inuit" which
aren't in the official 128 answer pool). Chapter 1's page 16 map labels are
transcribed, including its two Massachusetts labels and Atlantic Ocean.
These are text editions, not facsimiles. Chapters 7, 8, and 10 cover
difficult periods of the source material — Native American depopulation,
the start of slavery, colonial-era warfare, and the Civil War — factually,
in the source's own words, without added commentary. Chapter 10's Frederick
Douglass has no dedicated question in the official 128 and is covered
anyway, per this project's second goal of teaching the material, not only
the test answers.

Chapters 1, 6, 7, and 8 additionally each carry a genuinely public-domain
reference map sourced from outside the four supplied PDFs — the only
exception to "the PDFs are the source of truth" — added because the Study
Guide's own map graphics in these chapters carry no printed credit and so
cannot themselves be reproduced. Each candidate was found via search, then
its public-domain status was independently confirmed on the actual source
page before download; this is explicitly not the same thing as a "fair use"
image search, which was considered and rejected as too legally thin for this
project. `catalog.go` ties one manifest image to exactly one chapter page, so
the same National Archives 1775 colonies map is copied under four separate
manifest IDs/files (one per chapter) rather than shared. See
`internal/content/data/images/EXTERNAL-SOURCES.txt` for the exact source URL,
quoted license statement, and verification date of each. These added images
supplement rather than replace the exact facts (ratification dates, territory
names) their chapters' text transcriptions already carry, since the sourced
maps don't show that source-specific information. Two categories researched
for this effort — a territories-with-capitals map, and a historical map
showing both colonization and slave-trade Atlantic routes for Chapter 7 —
had no verifiable public-domain candidate and remain text only.

Question links are authored from topics actually covered, and a question whose
content is genuinely repeated across chapters (e.g. Q18, Q41, Q13, Q50, Q64,
Q81, Q86, Q14, Q82, Q58) links to all of them — Chapter 9 alone adds three new
triple/dual links (Q86 now spans Chapters 6, 8, and 9; Q14 spans Chapters 5
and 9; Q82 spans Chapters 1 and 9) by restating the Constitutional
Convention, Federalist Papers, and George Washington content already
covered elsewhere, in more depth. Chapter 10 introduces no new multi-chapter
links: its Civil War/Lincoln/Emancipation/Juneteenth/women's-rights-movement
content is new ground each of its seven linked questions covers only there.
Chapter 11 adds one new dual link (Q58, the federal government's exclusive
power to declare war: Chapter 1 states this generally, Chapter 11 narrates
Congress actually exercising it in 1941); its Korean War and Vietnam War
mentions are listed without "why" reasoning, so the corresponding questions
are deliberately not linked here. Chapter 12 adds five multi-chapter links
by restating already-covered facts with new specifics: Q86 (George
Washington's "Father of Our Country" epithet, now spanning Chapters 6, 8,
9, and 12), Q91 (the War of 1812 and the Civil War as wars fought in the
1800s, joining Chapter 9's Mexican-American War statement of the same
phrasing pattern), Q122 (the 50 stars, joining Chapter 6's states list),
Q128 (Veterans Day, joining Chapter 11's introduction), and Q126 (national
holidays, now spanning Chapters 10, 11, and 12, the last being the
definitive source since it carries the full "U.S. Holidays" list). Chapter
12 also adds one new single-chapter link, Q125 ("What is Independence
Day?"), stated directly by its own Independence Day section; Q8 ("Why is
the Declaration of Independence important?") is deliberately not linked
despite similar wording, to stay consistent with Chapter 8, which states
the same fact without being linked to Q8 either. Questions without a
published chapter do not get speculative links. Chapters 1–12 cover 25, 20,
12, 8, 14, 5, 4, 11, 8, 7, 15, and 11 questions respectively. All 12
chapters are now published, but the all-128 chapter-coverage requirement
remains open: 107 of 128 official questions are linked to a chapter whose
own prose states an accepted answer; the other 21 (e.g. Q1, Q12, Q55, Q87,
Q97, the "why did the U.S. enter war X" questions for wars only named in
passing) are not stated verbatim anywhere in the 12 chapters and are
deliberately left unlinked rather than forced — this project links only
what a chapter's own text actually says.
