.PHONY: default vet test bench fuzz doc doc-bench pre-release build-scripts

PATTERN=.
TEST_PATH=./...
MIN_COVERAGE=99

default: vet test

test:
	if ( which gotest &> /dev/null ); then \
		gotest -v -race -run=${PATTERN} -covermode=atomic -coverprofile=coverage.out ${TEST_PATH}; \
	else \
		go test -v -race -run=${PATTERN} -covermode=atomic -coverprofile=coverage.out ${TEST_PATH}; \
	fi
	go tool cover -html=coverage.out -o coverage.html
	@echo
	@cat coverage.out | awk '/\s0$$/{notCovered += 1} !/\s0$$/{covered += 1} END{coverPerc=covered / (covered + notCovered) * 100; printf("final coverage report: covered=%d, notCovered=%d, coverage=%.2f%%\n", covered, notCovered, coverPerc); if (coverPerc < ${MIN_COVERAGE}){printf("coverage bellow min acceptable of %.1f%%\n", ${MIN_COVERAGE}); exit 1}}'

vet:
	go vet ./...
	@# staticcheck (needs: go install honnef.co/go/tools/cmd/staticcheck@latest)
	staticcheck ./...
	@# errcheck (needs: go install github.com/kisielk/errcheck@latest)
	errcheck ./...

bench:
	go test -bench=. -benchtime=5s -benchmem ./... | tee docs/bench.txt

fuzz:
	go test -fuzz=FuzzOMapImpls ./omap/

build-scripts:
	$(MAKE) -C scripts/

doc: doc-bench

doc-bench: build-scripts
	./scripts/scripts benchtable docs/bench.txt docs/benchmarks.md

pre-release: vet test doc
	@echo
	@echo "Good to Go!!!"
	@echo
