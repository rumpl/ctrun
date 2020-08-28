all: cli

cli:
	@go build -o /out/ctrun .

.PHONY: cli
