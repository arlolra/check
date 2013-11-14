SHELL := /bin/bash

start: exits i18n
	@./check

rsync_server = metrics.torproject.org
consensuses_dir = metrics-recent/relay-descriptors/consensuses/
exit_lists_dir = metrics-recent/exit-lists/
descriptors_dir = metrics-recent/relay-descriptors/server-descriptors/

data/:
	@mkdir -p data

data/descriptors/: data/
	@mkdir -p data/descriptors

data/consensuses/: data/
	@mkdir -p data/consensuses

data/exit-lists/: data/
	@mkdir -p data/exit-lists

data/consensus: data/consensuses/
	@echo Getting latest consensus documents
	@rsync -avz $(rsync_server)::$(consensuses_dir) --delete ./data/consensuses/
	@echo Consensuses written

data/exit-addresses: data/exit-lists/
	@echo Getting latest exit lists
	@rsync -avz $(rsync_server)::$(exit_lists_dir) --delete ./data/exit-lists/
	@echo Exit lists written

exits: data/consensus data/exit-addresses data/cached-descriptors
	@echo Generating exit-policies file
	@python scripts/exitips.py
	@echo Done

data/cached-descriptors: descriptors
	@echo "Concatenating data/descriptors/* into data/cached-descriptors"
	@rm -f data/cached-descriptors
	find data/descriptors -type f -mmin -60 | xargs cat > data/cached-descriptors
	@echo "Done"

descriptors_cutoff = $(shell date -v-1H -v-30M "+%Y/%m/%d %H:%M:%S")
descriptors: data/descriptors/
	@echo "Getting latest descriptors (This may take a while)"
	@find data/descriptors -type f -mmin +90 -delete
	@rsync $(rsync_server)::$(descriptors_dir) | awk 'BEGIN { before="$(descriptors_cutoff)"; } before < ($$3 " " $$4) && ($$5 != ".") { print $$5; }' | rsync -avz --files-from=- $(rsync_server)::$(descriptors_dir) ./data/descriptors/
	@echo Done

data/langs: data/
	curl -k https://www.transifex.com/api/2/languages/ > data/langs

build:
	go fmt
	go build

# Add -i for installing latest version, -v for verbose
test: build
	go test -v -run "$(filter)"

cover: build
	go test -coverprofile cover.out

filter?=.
bench: build
	go test -i
	go test -benchtime 10s -bench "$(filter)" -benchmem

profile: build
	go test -cpuprofile ../../cpu.prof -memprofile ../../mem.prof -benchtime 40s -bench "$(filter)"

i18n: locale/ data/langs

locale/:
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

install: build
	mv check /usr/local/bin/check
	cp scripts/check.init /etc/init.d/check
	update-rc.d check defaults

.PHONY: start build i18n exits test bench cover profile descriptors install