SOURCES=$(wildcard *.coffee)
PRODUCTS=$(SOURCES:%.coffee=%.js)

.PHONY: all clean

all: common.js home.js login.js

common.js:
	coffee -c api.coffee
	cat api.js >common.js
	rm api.js

home.js:
	coffee -c home.coffee

login.js:
	coffee -c login.coffee

clean:
	$(RM) common.js home.js login.js