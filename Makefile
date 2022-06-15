.PHONY: default vet test bench fuzz doc doc-api doc-bench pre-release

default: vet test

test:
	go test -v -race -covermode=atomic -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

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
