NAME=bopher
DEPEND=github.com/Masterminds/glide

.PHONY: depend clean build

clean-build: clean depend build

build: depend
	go build github.com/kaakaa/bopher/bopher

depend:
	go get -v $(DEPEND)
	glide install

clean:
	rm -fr bopher.exe vendor/*
	touch vendor/.gitkeep
