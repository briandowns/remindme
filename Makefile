GO = go

BINDIR := bin
BINARY := remindme

VERSION = v1.1.1
GIT_SHA = $(shell git rev-parse HEAD)
LDFLAGS = -ldflags '-w -s -extldflags "-static" -X main.gitSHA=$(GIT_SHA) -X main.version=$(VERSION) -X main.name=$(BINARY)'

TAGS    = "netgo osusergo no_stage static_build"
$(BINDIR)/$(BINARY): clean
	$(GO) build -tags $(TAGS) -v $(LDFLAGS) -o $@

.PHONY: clean
clean:
	$(GO) clean
	rm -f $(BINDIR)/$(BINARY)

.PHONY: image
image: $(BINDIR)/$(BINARY)
	docker build -t briandowns/$(BINARY):$(VERSION) .
