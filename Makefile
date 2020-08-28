export DOCKER_BUILDKIT=1

all: cli

cli:
	@docker build . --target cli \
	--platform local \
	--output ./bin

.PHONY: cli
