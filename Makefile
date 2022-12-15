.PHONY: test
test:
	go test -race -count 1 ./...

.PHONY: supertest
supertest:
	go test -race -count 100 ./...

.PHONY: bench
bench:
	go test -bench=. -benchmem

.PHONY: lint
lint:
	golangci-lint run ./...