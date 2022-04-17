
.PHONY: all
all: install-git-precommit fmt test lint bin

.PHONY: precommit
precommit: install-git-precommit fmt test lint

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
	mkdir -p dist/{linux,darwin}/{amd64,arm64}/
	GOOS=linux GOARCH=amd64 go build -o dist/linux/amd64/lego-dnsserver .
	GOOS=linux GOARCH=arm64 go build -o dist/linux/arm64/lego-dnsserver .
	GOOS=darwin GOARCH=amd64 go build -o dist/darwin/amd64/lego-dnsserver .
	GOOS=darwin GOARCH=arm64 go build -o dist/darwin/arm64/lego-dnsserver .

.PHONY: install-git-precommit
install-git-precommit:
	cp hacks/githook-pre-commit.sh .git/hooks/pre-commit
