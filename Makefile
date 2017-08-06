.PHONY: all build run shell

# Note that this requires you have the directory checked out to /tmp/maps
MAPS=-v /tmp/maps:/src/github.com/jemoster/icfp2017/maps
PLAYLOG=-v /tmp/playlogs:/src/github.com/jemoster/icfp2017/data
VOLS=$(MAPS) $(PLAYLOG)
RUNNER=./tools/bot_runner/online_adapter.py
PLAYER=./tools/bot_runner/make_player_data.py
IDLERUN=./tools/bot_runner/idle_runner.py
IDLENAME=idle-tmp

all: run

build: Dockerfile
	docker build -t boxes .

run: build
	docker run $(VOLS) --rm boxes

shell: build
	docker run $(VOLS) --rm --entrypoint bash -it boxes

simpleton-run: build
	docker run $(VOLS) --rm boxes $(RUNNER) ./simpleton 9017

walkbot-run: build
	docker run $(VOLS) --rm boxes $(RUNNER) ./walk 9200

brownian-run: build
	docker run $(VOLS) --rm boxes $(RUNNER) ./brownian 9196

blob-run: build
	docker run $(VOLS) --rm boxes $(RUNNER) ./blob 9196

punter2-run: build
	docker run $(VOLS) --rm boxes $(RUNNER) ./punter2 9005

random-run: build
	docker run $(VOLS) --rm boxes $(RUNNER) ./src/pybots/ai_random.py 9018

trollbot-run: build
	docker run $(VOLS) --rm boxes $(RUNNER) ./src/pybots/trollbot.py 9231

robbrown-run: build
	docker run $(VOLS) --rm boxes $(RUNNER) ./src/pybots/ai_olrobbrown.py 9231

troll-run: build
	docker run $(VOLS) --rm boxes $(RUNNER) ./src/pybots/trollbot.py 9238

multiplay: build
	docker run $(VOLS) --rm boxes $(PLAYER) 3 9234

idle-run: build
	docker run $(VOLS) --name $(IDLENAME) -d boxes $(IDLERUN)
