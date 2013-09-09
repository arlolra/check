SHELL := /bin/bash

export GOPATH := $(CURDIR):$(GOPATH)
export TORCHECKBASE := $(CURDIR)/

start: data/exit-policies data/langs
	@./check

# Get any data files we're missing
# Be careful with trailing whitespace on these lines
rsync_server = metrics.torproject.org
consensuses_dir = metrics-recent/relay-descriptors/consensuses/
exit_lists_dir = metrics-recent/exit-lists/
latest_exit_list = $(shell rsync $(rsync_server)::$(exit_lists_dir) | tail -1 | tr " " "\n" | tail -1)
latest_consensus = $(shell rsync $(rsync_server)::$(consensuses_dir) | tail -1 | tr " " "\n" | tail -1)

data/:
	@mkdir -p data

public/:
	@mkdir -p public

data/consensus: data/
	@echo Getting latest consensus document
	@rsync $(rsync_server)::$(consensuses_dir)$(strip $(latest_consensus)) ./data/consensus
	@echo Consensus written to file

public/exit-addresses: public/
	@echo Getting latest exit lists
	@rsync $(rsync_server)::$(exit_lists_dir)$(strip $(latest_exit_list)) ./public/exit-addresses
	@echo Exit lists written to file

data/exit-policies: data/consensus public/exit-addresses
	@echo Generating exit-policies file
	@python scripts/exitips.py
	@echo Done

data/langs: data/
	curl https://www.transifex.com/api/2/languages/ > data/langs

build:
	go fmt ./src/check
	go fmt
	go build

# Add -i for installing latest version, -v for verbose
test: build
	go test check -v -run "$(filter)"

cover: build
	go test check -coverprofile cover.out

filter?=.
bench: build
	go test check -i 
	go test check -benchtime 10s -bench "$(filter)" -benchmem

profile: build
	go test check -cpuprofile ../../cpu.prof -memprofile ../../mem.prof -benchtime 40s -bench "$(filter)"

i18n:
	rm -rf locale
	git clone -b torcheck https://git.torproject.org/translation.git locale
	pushd locale; \
	for f in *; do \
		if [ "$$f" != "templates" ]; then \
			pushd "$$f"; \
			mkdir LC_MESSAGES; \
			msgfmt -o LC_MESSAGES/check.mo torcheck.po; \
			popd; \
		fi \
	done

.PHONY: start build i18n test bench coverage profile
