.PHONY: all build run shell

# Note that this requires you have the directory checked out to /tmp/maps
MAPS=-v /tmp/maps:/app/github.com/jemoster/icfp2017/maps
RUNNER=python3 -u tools/bot_runner/online_adapter.py

all: run

build: Dockerfile
	docker build -t boxes .

run: build
	docker run $(MAPS) --rm boxes

shell: build
	docker run $(MAPS) --rm --entrypoint bash -it boxes

simpleton-run: build
	docker run $(MAPS) boxes $(RUNNER) ./simpleton 9017

walkbot-run: build
	docker run $(MAPS) boxes $(RUNNER) ./walk 9017
