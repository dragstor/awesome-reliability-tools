.PHONY: all fmt test validate generate check clean favicons favicons-refresh

all: generate

fmt:
	gofmt -w cmd internal

test:
	go test ./...

validate:
	go run ./cmd/validate

generate:
	go run ./cmd/generate

# favicons is the only networking command. It is intentionally separate from
# generate so that normal builds stay deterministic and offline.
favicons:
	go run ./cmd/favicons

favicons-refresh:
	go run ./cmd/favicons --force

# check is the single gate used locally and in CI: it fails when the tracked
# generated artifact (README.md) is stale or when generation is not deterministic.
check: test validate generate
	@echo "checking tracked generated artifacts are up to date..."
	git diff --exit-code
	@echo "checking generation is deterministic..."
	@tmp=$$(mktemp -d) && \
		cp -r dist $$tmp/first && \
		$(MAKE) generate >/dev/null && \
		cp -r dist $$tmp/second && \
		diff -r $$tmp/first $$tmp/second && \
		rm -rf $$tmp
	@echo "ok"

clean:
	rm -rf dist
