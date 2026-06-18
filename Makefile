GO := go

build:
	@CGO_ENABLED=0 $(GO) build -o $(CURDIR)/bin/

run: build
	@$(CURDIR)/bin/tunnel7
