---
{
  "id": "patriotic-symbols",
  "title": "Patriotic Symbols of the United States",
  "kind": "symbol",
  "source": "citizens-almanac",
  "citation": "U.S. Department of Homeland Security, U.S. Citizenship and Immigration Services, Office of Citizenship, The Citizen's Almanac, Washington, DC, 2014.",
  "questions": [66, 121, 122, 124],
  "source_start": 20,
  "source_end": 26,
  "editorial_notes": [
    "This document completes M-76's 'Patriotic Anthems and Symbols of the United States' section (pages 9-26) alongside patriotic-anthems.md: the section's other four items (Pledge of Allegiance, Flag, Motto, Great Seal), covering pages 20-26 so both documents' own page ranges stay fully contiguous, since TestLibrarySourceCoverage requires every page within a document's own source_start/source_end to be accounted for.",
    "Each item's decorative cursive title ('Pledge of Allegiance,' 'Flag of the United States of America,' 'Motto of the United States,' 'Great Seal of the United States') renders as pure vector art with no extractable PDF text layer at all, the same gap patriotic-anthems.md's song titles have; each is recorded in data/raw/almanac-visual-supplement.json and authored here as a real heading, not decoration to omit.",
    "The Almanac's two-column layout hyphenates words across a line wrap on nearly every one of these seven pages, several landing on a row that also carries the other column's or an image caption's unrelated text; each is fixed as one exact block in TestLibrarySourceCoverage's almanacTextFixes map, the same technique patriotic-anthems.md's fixes use.",
    "One sentence splits across a page boundary in the source itself, authored as a matching split across two blocks: '...as part of official Columbus' (page 20) / 'Day observances to celebrate...' (page 21).",
    "Q66 ('What do we show loyalty to when we say the Pledge of Allegiance?') is linked: the Pledge itself names 'the Flag of the United States of America' and 'the Republic for which it stands.' Q121 (13 stripes) and Q122 (50 stars) are linked: the Flag section states the stripes and stars honor the thirteen original states and that the flag has 'one star for every state.' Q124 (the meaning of 'E Pluribus Unum') is linked from the Great Seal section, which is the one page here that actually translates the phrase ('out of many, one'); the Motto section names E Pluribus Unum as the Nation's original 1776 motto but never translates it, so the link belongs to the Great Seal page, not the Motto page, consistent with this project's practice of linking only where a fact is actually stated. No official question asks about the Motto or the Great Seal directly; both are included anyway, per this project's second goal of teaching the material, not only the test answers."
  ]
}
---
<!-- page:20 -->
### Pledge of Allegiance

The Pledge of Allegiance was first published on September 8, 1892, in the Youth's Companion magazine. The original pledge read as follows, "I pledge allegiance to my Flag and the Republic for which it stands: one Nation indivisible, with Liberty and Justice for all." Children in public schools across the country recited the pledge for the first time on October 12, 1892, as part of official Columbus

> Students recite the Pledge of Allegiance in a Washington, DC, classroom (ca. 1899). Courtesy of the Library of Congress, LC-USZ62-14693

> Citizens of Vale, OR, take off their hats during the Pledge of Allegiance, July 4, 1941. Courtesy of the Library of Congress, LC-USF33-013070-M2

<!-- page:21 -->
Day observances to celebrate the 400th anniversary of his discovery of America.

In 1942, by an official act, Congress recognized the pledge. The phrase "under God" was added to the pledge by another act of Congress on June 14, 1954. Upon signing the legislation to authorize the addition, President Dwight D. Eisenhower said, "In this way we are reaffirming the transcendence of religious faith in America's heritage and future; in this way we shall constantly strengthen those spiritual weapons which forever will be our country's most powerful resource in peace and war."

When delivering the Pledge of Allegiance, all must be standing at attention, facing the flag with the right hand over the heart. Men not in uniform should remove any nonreligious headdress with their right hand and hold it at the left shoulder, the hand being over the heart. Those in uniform should remain silent, face the flag, and render the military salute.

```verse
Pledge of Allegiance

I pledge allegiance
to the Flag
of the United States of America,
and to the Republic
for which it stands,
one Nation
under God,
indivisible,
with liberty
and justice for all.
```

<!-- page:22 -->
### Flag of the United States of America

As America fought for its independence from Great Britain, it soon became evident that the new nation needed a flag of its own to identify American forts and ships.

> The flag that was authorized by Congress on June 14, 1777.

A design of thirteen alternating red and white stripes and thirteen stars in a blue field was accepted by the Continental Congress on June 14, 1777. These stars and stripes honored the thirteen states that had joined together to form the United States of America.

As the United States expanded, however, more states were added to the Union. To celebrate the Nation's growth, Congress decided that the flag should become a visible symbol of change and established that the American flag would have one star for every state. The design of the American flag has changed twenty-seven times, and since 1959 it has had fifty stars and thirteen stripes.

> The U.S. flag today.

The American flag is called the "Star-Spangled Banner," the "Stars and Stripes," the "Red, White, and Blue," and "Old Glory." To emphasize the importance of the American flag to the Nation and its people, Congress established June 14 of each year as Flag Day. On this day, Americans take special notice of the flag and reflect on its meaning.

<!-- page:23 -->
### Motto of the United States

On July 30, 1956, President Dwight D. Eisenhower approved a Joint Resolution of the 84th Congress officially establishing the phrase, "In God We Trust," as the national motto of the United States. "In God We Trust" replaced the phrase, E Pluribus Unum, which had been selected as the Nation's official motto in 1776.

The motto, "In God We Trust," can be traced back nearly 200 years in U.S. history. During the War of 1812, as the morning light revealed that the American flag was still waving above Fort McHenry, Francis Scott Key wrote the poem that would eventually become our national anthem. The final stanza of the poem read, "And this be our motto: 'In God is our trust!'" In 1864, Key's phrase was changed to "In God We Trust" and included on the redesigned two-cent coin. The following year, Congress authorized the Director of the Philadelphia Mint to place the motto on all gold and silver coins. The motto began appearing on all U.S. coins in 1938. "In God We Trust" became a part of the design of U.S. currency (paper money) in 1957. The Bureau of Engraving and Printing has incorporated the motto on all currency since 1963.

"In God We Trust" is also engraved on the wall above the Speaker's dais in the Chamber of the House of Representatives and over the entrance to the Chamber of the Senate.

> President Lyndon B. Johnson delivering his State of the Union address before a joint session of Congress, January 8, 1964. Engraved above the Speaker's dais is the motto "In God We Trust." Courtesy of the Lyndon Baines Johnson Library and Museum

<!-- page:24 -->
### Great Seal of the United States

On July 4, 1776, the Continental Congress appointed a committee to create a seal for the United States of America. Following the appointment of two additional committees, each building upon the other, the Great Seal was finalized and approved on June 20, 1782.

The Great Seal has two sides—an obverse, or front side, and a reverse side. The obverse side displays a bald eagle, the national bird, in the center. The bald eagle holds a scroll inscribed E pluribus unum in its beak. The phrase means "out of many, one" in Latin and signifies one nation that was created from thirteen separate colonies. In one of the eagle's claws is an olive branch and in the other is a bundle of thirteen arrows. The olive branch signifies peace and the arrows signify war.

> Obverse side of the Great Seal of the United States. Courtesy of the U.S. Department of State

<!-- page:25 -->
> Reverse side of the Great Seal of the United States. Courtesy of the U.S. Department of State

A shield with thirteen red and white stripes covers the eagle's breast. The eagle alone supports the shield to signify that Americans should rely on their own virtue and not that of other nations. The red and white stripes of the shield represent the states united under and supporting the blue, representing the President and Congress. The color red signifies valor and bravery, the color white signifies purity and innocence, and the color blue signifies vigilance, perseverance, and justice. Above the eagle's head is a cloud that surrounds a blue field containing thirteen stars, which form a constellation. The constellation represents the fact that the new Nation is taking its place among the sovereign powers.

The reverse side contains a thirteen-step pyramid with the year 1776 in Roman numerals at its base. Above the pyramid is the Eye of Providence and the motto Annuit Coeptis, meaning "He [God] favors our undertakings." Below the pyramid, Novus Ordo Seclorum, meaning "New Order of the Ages," is written on a scroll to

<!-- page:26 -->
signify the beginning of the new American era.

The obverse side of the Great Seal is used on postage stamps, military uniforms, U.S. passports, and above the doors of U.S. embassies worldwide. Both sides are present on the one dollar bill.
