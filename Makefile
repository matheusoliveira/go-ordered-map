.PHONY: default vet test bench fuzz doc doc-api doc-bench pre-release

PATTERN=.
TEST_PATH=./...
MIN_COVERAGE=99

default: vet test

test:
	go test -v -race -run=${PATTERN} -covermode=atomic -coverprofile=coverage.out ${TEST_PATH}
	go tool cover -html=coverage.out -o coverage.html
	@echo
	@cat coverage.out | awk '/\s0$$/{notCovered += 1} !/\s0$$/{covered += 1} END{coverPerc=covered / (covered + notCovered) * 100; printf("final coverage report: covered=%d, notCovered=%d, coverage=%.2f%%\n", covered, notCovered, coverPerc); if (coverPerc < ${MIN_COVERAGE}){printf("coverage bellow min acceptable of %.1f%%\n", ${MIN_COVERAGE}); exit 1}}'

vet:
	go vet ./...
	# staticcheck (needs: go install honnef.co/go/tools/cmd/staticcheck@latest)
	!( which staticcheck &> /dev/null ) || staticcheck ./...

bench:
	go test -bench=. -benchtime=5s -benchmem ./... | tee docs/bench.txt

fuzz:
	go test -fuzz=FuzzOMapImpls ./pkg/omap/

doc: doc-api doc-bench

doc-api:
	go run github.com/robertkrimen/godocdown/godocdown@latest ./pkg/omap/ \
		| sed -E 's#^(\s+) import "."#\n```go\nimport "github.com/matheusoliveira/go-ordered-map/pkg/omap"\n```#g' \
		| grep -v '^--$$' \
		> docs/api.md

doc-bench:
	go run utilities/benchtable.go

pre-release: vet test doc
