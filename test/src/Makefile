CFLAGS = -g
INSTFLAGS = -finstrument-functions
INSTLIBS = -lauklet

all: release install

snellius-debug: snellius.c
	gcc -o $@ $< ${CFLAGS} ${INSTFLAGS} ${INSTLIBS}

snellius: snellius-debug
	cp $< $@
	strip $@

release: snellius snellius-debug
	. ./secrets.sh &&\
	export AUKLET_RELEASE_ENDPOINT=staging &&\
	releaser -appid $$APP_ID -apikey $$API_KEY -debug snellius-debug -deploy snellius

install: snellius
	sudo cp $< /usr/local/bin/

clean:
	rm -f snellius-debug snellius

.PHONY: all clean release install
