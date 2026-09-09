# Unified testing contract (see .claude/skills/solution-conformance-testing).
# Four caller-facing targets, four caller-facing toggles:
#   WITH_CODE_COVERAGE=true   also emit the normalized coverage result + badge
#   ONLY_DELTA=true DELTA_BASE=<ref>   mutate only code changed since <ref>
#
# Everything stack-specific (Go, gremlins, gotestsum) stays inside this file.

SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

export CGO_ENABLED := 0

GO        ?= go
COVERPKG  ?= ./client/eaxmi,./internal/mcpserver,./internal/service/sparx
TESTKIT   := $(GO) run ./tools/testkit
GOTESTSUM ?= $(GO) run gotest.tools/gotestsum@v1.13.0
GREMLINS  ?= $(GO) run github.com/go-gremlins/gremlins/cmd/gremlins@v0.6.0

WITH_CODE_COVERAGE ?= false
ONLY_DELTA         ?= false
DELTA_BASE         ?=

# Packages gremlins mutates, one `gremlins unleash` run each (it accepts one
# path per invocation). Kept to the production packages that carry logic.
# --integration + --workers 1 + --coverpkg <pkg> (below): the Cucumber suite for
# a package lives in a sibling `test/` package, so a plain per-package coverage
# run sees the code as uncovered. --integration runs the whole test binary per
# mutant and --coverpkg attributes coverage back to <pkg>. --workers 1 + the
# large timeout coefficient avoid gremlins' adaptive-timeout collapse, which
# mislabels healthy mutants "TIMED OUT" on a fast parallel run.
MUTATE_PKGS ?= ./client/eaxmi ./internal/mcpserver ./internal/service/sparx

# gremlins derives each mutant's timeout from the baseline test duration; on a
# sub-second suite that estimate is far too tight and healthy mutants get
# mislabelled "TIMED OUT". A generous coefficient fixes it without hiding a
# genuinely hanging mutant (gremlins still caps the absolute wait).
MUTATE_TIMEOUT_COEFF ?= 60

.PHONY: unit-test mutation-test test-report test-and-report clean

## Run every test — Cucumber scenarios and plain Go tests — in one invocation,
## always with coverage. No -race: the codebase is pure Go with no
## goroutines in production code (the race detector also needs cgo).
unit-test:
	@mkdir -p tmp/result tmp/report/tests tmp/report/coverage
	@rm -f tmp/coverage.out
	set +e; \
	$(GOTESTSUM) --format testname \
		--junitfile tmp/report/tests/junit.xml \
		--jsonfile tmp/report/tests/go-test.json \
		-- ./... -coverpkg=$(COVERPKG) -coverprofile=tmp/coverage.out; \
	status=$$?; \
	$(TESTKIT) unit-test tmp/report/tests/go-test.json; \
	$(GO) tool cover -func=tmp/coverage.out | tee tmp/report/coverage/summary.txt; \
	$(GO) tool cover -html=tmp/coverage.out -o tmp/report/coverage/index.html; \
	if [ "$(WITH_CODE_COVERAGE)" = "true" ]; then \
		$(TESTKIT) coverage tmp/report/coverage/summary.txt; \
	fi; \
	exit $$status

## Run mutation testing with gremlins. Exits with gremlins' own status after
## writing the normalized result.
mutation-test:
	@mkdir -p tmp/result tmp/report/mutation
	@rm -f tmp/report/mutation/*.json
	set +e; \
	diffflag=""; \
	if [ "$(ONLY_DELTA)" = "true" ]; then \
		if [ -z "$(DELTA_BASE)" ]; then echo "ONLY_DELTA=true requires DELTA_BASE=<ref>" >&2; exit 2; fi; \
		diffflag="--diff $(DELTA_BASE)"; \
	fi; \
	status=0; \
	for pkg in $(MUTATE_PKGS); do \
		name=$$(echo "$$pkg" | sed 's#[./]#_#g;s#^_*##'); \
		$(GREMLINS) unleash "$$pkg" \
			--integration --workers 1 --timeout-coefficient $(MUTATE_TIMEOUT_COEFF) \
			--coverpkg "$$pkg" \
			--output "tmp/report/mutation/$$name.json" \
			$$diffflag || status=$$?; \
	done; \
	$(TESTKIT) mutation tmp/report/mutation/*.json; \
	exit $$status

## Assemble the readable public/ report from whatever the *-test targets left in
## tmp/. Reads only tmp/result/*.json — never a tool's native format.
test-report:
	$(TESTKIT) report

## Full sequence: unit-test (with coverage) -> mutation-test -> test-report.
test-and-report:
	$(MAKE) unit-test WITH_CODE_COVERAGE=true
	-$(MAKE) mutation-test ONLY_DELTA=$(ONLY_DELTA) DELTA_BASE=$(DELTA_BASE)
	$(MAKE) test-report

clean:
	rm -rf tmp public
