SHELL := /bin/bash

fresh: clean start

start: check
	@./check

check: 
	@echo "Generating new build"
	go fmt *.go
	go build *.go

build: clean check

clean: 
	@(rm ./check 2&>1) || true 

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

.PHONY: start build i18n clean
