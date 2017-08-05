.PHONY: all build run shell

MAPS=-v /tmp/maps:/app/github.com/jemoster/icfp2017/maps

all: run

build: Dockerfile
	docker build -t boxes .

run: build
	docker run $(MAPS) --rm boxes

shell: build
	docker run $(MAPS) --rm --entrypoint bash -it boxes
