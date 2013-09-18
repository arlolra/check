SHELL := /bin/bash

export GOPATH := $(CURDIR):$(GOPATH)
export TORCHECKBASE := $(CURDIR)/

start: data/exit-policies data/langs locale/
	@./check

# Get any data files we're missing
# Be careful with trailing whitespace on these lines
rsync_server = metrics.torproject.org
consensuses_dir = metrics-recent/relay-descriptors/consensuses/
exit_lists_dir = metrics-recent/exit-lists/
latest_exit_list = $(shell rsync $(rsync_server)::$(exit_lists_dir) | tail -1 | tr " " "\n" | tail -1)
latest_consensus = $(shell rsync $(rsync_server)::$(consensuses_dir) | tail -1 | tr " " "\n" | tail -1)
descriptors_dir = metrics-recent/relay-descriptors/server-descriptors/

data/:
	@mkdir -p data

data/descriptors/: data/
	@mkdir -p data/descriptors

locale/: i18n

data/consensus: data/
	@echo Getting latest consensus document
	@rsync $(rsync_server)::$(consensuses_dir)$(strip $(latest_consensus)) ./data/consensus
	@echo Consensus written to file

data/exit-addresses: data/
	@echo Getting latest exit lists
	@rsync $(rsync_server)::$(exit_lists_dir)$(strip $(latest_exit_list)) ./data/exit-addresses
	@echo Exit lists written to file

data/exit-policies: data/consensus data/exit-addresses data/all_descriptors
	@echo Generating exit-policies file
	@python scripts/exitips.py
	@echo Done

data/all_descriptors: data/ descriptors
	@echo "Concatenating data/descriptors/* into data/all_descriptors"
	@cat data/descriptors/0* > data/all_descriptors
	@cat data/descriptors/1* >> data/all_descriptors
	@cat data/descriptors/2* >> data/all_descriptors
	@cat data/descriptors/3* >> data/all_descriptors
	@cat data/descriptors/4* >> data/all_descriptors
	@cat data/descriptors/5* >> data/all_descriptors
	@cat data/descriptors/6* >> data/all_descriptors
	@cat data/descriptors/7* >> data/all_descriptors
	@cat data/descriptors/8* >> data/all_descriptors
	@cat data/descriptors/9* >> data/all_descriptors
	@cat data/descriptors/a* >> data/all_descriptors
	@cat data/descriptors/b* >> data/all_descriptors
	@cat data/descriptors/c* >> data/all_descriptors
	@cat data/descriptors/d* >> data/all_descriptors
	@cat data/descriptors/e* >> data/all_descriptors
	@cat data/descriptors/f* >> data/all_descriptors
	@echo "Done"

descriptors: data/descriptors/
	@echo "Getting latest descriptors (This may take a while)"
	rsync -avz $(rsync_server)::$(descriptors_dir) --delete ./data/descriptors/
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
	git clone -b torcheck_completed https://git.torproject.org/translation.git locale
	pushd locale; \
	for f in *; do \
		if [ "$$f" != "templates" ]; then \
			pushd "$$f"; \
			mkdir LC_MESSAGES; \
			msgfmt -o LC_MESSAGES/check.mo torcheck.po; \
			popd; \
		fi \
	done

.PHONY: start build i18n test bench cover profile descriptors
