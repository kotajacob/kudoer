# kudoer
# See LICENSE for copyright and license details.
.POSIX:

PREFIX ?= /usr
GO ?= go
GOFLAGS ?= -buildvcs=false
RM ?= rm -f

all: kudoer

kudoer:
	$(GO) build $(GOFLAGS) .

install: all
	mkdir -p $(DESTDIR)$(PREFIX)/bin
	cp -f kudoer $(DESTDIR)$(PREFIX)/bin
	chmod 755 $(DESTDIR)$(PREFIX)/bin/kudoer

uninstall:
	$(RM) $(DESTDIR)$(PREFIX)/bin/kudoer

clean:
	$(RM) kudoer

run:
	go run -race .

watch:
	fd -e go -e tmpl | entr -rcs "go run -race ."

.PHONY: all kudoer install uninstall clean run watch
