D2 ?= d2

.PHONY: build test lint ingest build-all diagrams
build:
	CGO_ENABLED=0 go build -trimpath -o bin/n400 ./cmd/n400
test:
	go test -race ./...
lint:
	go vet ./...
	staticcheck ./...
ingest:
	go run ./cmd/ingest
build-all:
	@set -e; for os in linux darwin windows; do for arch in amd64 arm64; do suffix=; if [ "$$os" = windows ]; then suffix=.exe; fi; CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -trimpath -o bin/n400-$$os-$$arch$$suffix ./cmd/n400; done; done
diagrams:
	$(D2) --theme=0 internal/content/data/diagrams/lawmaking.d2 internal/web/static/diagrams/lawmaking-light.svg
	$(D2) --theme=200 internal/content/data/diagrams/lawmaking.d2 internal/web/static/diagrams/lawmaking-dark.svg
	$(D2) --theme=0 internal/content/data/diagrams/legislative-branches.d2 internal/web/static/diagrams/legislative-branches-light.svg
	$(D2) --theme=200 internal/content/data/diagrams/legislative-branches.d2 internal/web/static/diagrams/legislative-branches-dark.svg
	$(D2) --theme=0 internal/content/data/diagrams/house-membership.d2 internal/web/static/diagrams/house-membership-light.svg
	$(D2) --theme=200 internal/content/data/diagrams/house-membership.d2 internal/web/static/diagrams/house-membership-dark.svg
	$(D2) --theme=0 internal/content/data/diagrams/legislative-comparison.d2 internal/web/static/diagrams/legislative-comparison-light.svg
	$(D2) --theme=200 internal/content/data/diagrams/legislative-comparison.d2 internal/web/static/diagrams/legislative-comparison-dark.svg
	$(D2) --theme=0 internal/content/data/diagrams/government-branches.d2 internal/web/static/diagrams/government-branches-light.svg
	$(D2) --theme=200 internal/content/data/diagrams/government-branches.d2 internal/web/static/diagrams/government-branches-dark.svg
	$(D2) --theme=0 internal/content/data/diagrams/congress-two-parts.d2 internal/web/static/diagrams/congress-two-parts-light.svg
	$(D2) --theme=200 internal/content/data/diagrams/congress-two-parts.d2 internal/web/static/diagrams/congress-two-parts-dark.svg
	$(D2) --theme=0 internal/content/data/diagrams/separation-of-powers.d2 internal/web/static/diagrams/separation-of-powers-light.svg
	$(D2) --theme=200 internal/content/data/diagrams/separation-of-powers.d2 internal/web/static/diagrams/separation-of-powers-dark.svg
	$(D2) --theme=0 internal/content/data/diagrams/line-of-succession.d2 internal/web/static/diagrams/line-of-succession-light.svg
	$(D2) --theme=200 internal/content/data/diagrams/line-of-succession.d2 internal/web/static/diagrams/line-of-succession-dark.svg
	$(D2) --theme=0 internal/content/data/diagrams/federal-court-system.d2 internal/web/static/diagrams/federal-court-system-light.svg
	$(D2) --theme=200 internal/content/data/diagrams/federal-court-system.d2 internal/web/static/diagrams/federal-court-system-dark.svg
	$(D2) --theme=0 internal/content/data/diagrams/voting-rights-timeline.d2 internal/web/static/diagrams/voting-rights-timeline-light.svg
	$(D2) --theme=200 internal/content/data/diagrams/voting-rights-timeline.d2 internal/web/static/diagrams/voting-rights-timeline-dark.svg
	# D2's bundled dark theme is purple; map its emitted palette to the app's dark tokens.
	sed -i -e 's/#CBA6[fF]7/#f19c77/g' -e 's/#f38BA8/#f19c77/g' -e 's/#CDD6F4/#eeeadd/g' -e 's/#BAC2DE/#b4b2a3/g' -e 's/#A6ADC8/#b4b2a3/g' -e 's/#6C7086/#b4b2a3/g' -e 's/#585B70/#505147/g' -e 's/#45475A/#2b2c26/g' -e 's/#313244/#2b2c26/g' -e 's/#1E1E2E/#23231f/g' internal/web/static/diagrams/*-dark.svg
