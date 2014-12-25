SOURCES=$(wildcard *.coffee)
PRODUCTS=$(SOURCES:%.coffee=%.js)

.PHONY: all clean

all: output.js

output.js:
	coffee -c *.coffee
	cat *.js >output.js
	rm $(PRODUCTS)

clean:
	rm -f $(PRODUCTS) output.js