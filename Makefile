
.PHONY: all
all: git-prehook fmt test lint

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: test
test:
	go vet ./...
	go test ./...

.PHONY: install-golang-ci-lint
install-golang-ci-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.45.2

.PHONY: lint
lint: install-golang-ci-lint
	golangci-lint run

.PHONY: lint-fix
lint-fix: install-golang-ci-lint
	golangci-lint run --fix

.PHONY: bin
bin:
	mkdir -p dist/{linux}/{amd64}/
	GOOS=linux GOARCH=amd64 go build -o dist/linux/amd64/lego-dnsserver .

.PHONY: git-prehook
git-prehook:
	cp hacks/githook-pre-commit.sh .git/hooks/pre-commit
