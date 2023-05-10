ifeq ($(OS),Windows_NT)
	rmdir := rmdir /s /q
else
	rmdir := rm -rf
endif

GOCMD=GOOS=$(GOOS) GOARCH=$(GOARCH) go

all: test

test: clean
	mkdir -p .tests && \
	${GOCMD} test -coverpkg=. -coverprofile=cover.out -outputdir=.tests ./... | tee .tests/report.test && \
    ${GOCMD} tool test2json -t < .tests/report.test > .tests/report.json

clean:
	$(rmdir) .tests