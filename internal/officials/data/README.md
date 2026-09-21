# Dated official-answer data

`officials.json` is a bundled snapshot dated 2026-09-21. Each federal value
records the official source URL used for review in its `sources` object. It is
not a USCIS source; the web interface always retains the USCIS test-updates
verification link.

`states.json` contains the 50 states and District of Columbia, their official
USPS abbreviations, and capitals. State names and abbreviations were checked
against the U.S. Census Bureau ANSI code list. Capital names were checked
against the Minnesota Legislative Reference Library's state-capitals table.

The snapshot currently covers President, Vice President, Speaker of the House,
Chief Justice, state capitals, and two senators for every state. Governors,
representatives, ZIP district candidates, the live sidecar, and refresh clients
are intentionally not represented as blank answers; they remain unavailable
until their verified data sources are added.
