MMMBED ?= mmmbed
CC ?= cc

.PHONY: all check clean
all: app
asset.c: greeting.txt
	$(MMMBED) -l c --omit-parent --variable greeting $< $@
app: main.c asset.c
	$(CC) -o $@ main.c
check:
	$(MMMBED) --check -l c --omit-parent --variable greeting greeting.txt asset.c
clean:
	rm -f app asset.c
