all: debug deploy

debug: main.c
	gcc -o $@ -g $<

deploy: debug
	cp $< $@
	strip $@

clean:
	rm -f debug deploy
