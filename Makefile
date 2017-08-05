.PHONY: all build run shell

# Note that this requires you have the directory checked out to /tmp/maps
MAPS=-v /tmp/maps:/app/github.com/jemoster/icfp2017/maps
PLAYLOG=-v /tmp/playlogs:/app/github.com/jemoster/icfp2017/data
VOLS=$(MAPS) $(PLAYLOG)
RUNNER=python3 -u tools/bot_runner/online_adapter.py
PLAYER=./tools/bot_runner/make_player_data.py

all: run

build: Dockerfile
	docker build -t boxes .

run: build
	docker run $(MAPS) --rm boxes

shell: build
	docker run $(MAPS) --rm --entrypoint bash -it boxes

simpleton-run: build
	docker run $(MAPS) --rm boxes $(RUNNER) ./simpleton 9017

walkbot-run: build
	docker run $(MAPS) --rm boxes $(RUNNER) ./walk 9017

multiplay: build
	docker run $(MAPS) --rm boxes $(PLAYER) 4 9053
