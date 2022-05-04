
AUTOMAKE_OPTIONS = foreign no-dependencies no-installinfo

SUFFIXES = .go
OBJEXT= none

bin_PROGRAMS = dedup
dedup_SOURCES = dedup.go

man1_MANS = dedup.man

GOFLAGS = -ldflags="-s -w"

dedup$(EXEEXT): $(dedup_SOURCES)
	$(GO) build $(GOFLAGS) -o dedup$(EXEEXT) $(dedup_SOURCES)

EXTRA_DIST = \
	go.mod \
	go.sum

EXTRA_DIST += vendor/


clean-local:
	rm -rf autom4te.cache
	rm -rf log/ run/ data/
