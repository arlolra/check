SHELL := /bin/bash

start:
	@./check

build:
	go fmt check.go
	go build check.go

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

.PHONY: start build i18n
