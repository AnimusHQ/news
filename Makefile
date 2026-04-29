.PHONY: test vet fmt check validate dry-run

test:
	go test ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

check: fmt vet test

validate:
	go run ./cmd/animus-news validate-episode episodes/0001-after-git-push

dry-run:
	go run ./cmd/animus-news dry-run episodes/0001-after-git-push
