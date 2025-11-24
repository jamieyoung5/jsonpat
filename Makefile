.PHONY: test benchmark lint clean

test:
	go test -race -cover ./...

benchmark:
	go test -bench=. -benchmem ./...

lint:
	golangci-lint run

clean:
	go clean
	rm -f coverage.txt
