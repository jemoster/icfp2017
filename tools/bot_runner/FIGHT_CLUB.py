#!/usr/bin/env python3
import subprocess
import argparse
from time import sleep

# Note that this requires you have the directory checked out to /tmp/maps
BASEDIR = '/src/github.com/jemoster/icfp2017/'
HOSTDIR = 'c:/Users/Joe/Documents/src/icfp2017/'
# HOSTDIR = '/tmp/maps'

MAPS = '-v {HOSTDIR}maps:{BASEDIR}maps'.format(HOSTDIR=HOSTDIR, BASEDIR=BASEDIR)
PLAYLOG = '-v {}playlogs:{}data'.format(HOSTDIR, BASEDIR)

VOLS = '{} {}'.format(MAPS, PLAYLOG)
RUNNER = './tools/bot_runner/online_adapter.py'
PLAYER = './tools/bot_runner/make_player_data.py'
IDLERUN = './tools/bot_runner/idle_runner.py'
IDLENAME = 'idle-tmp'
SERVER = 'icfp_local'
NET = '--net {SERVER}'.format(SERVER=SERVER)


BOT_INDEX = {
    'simpleton': './simpleton',
    'walkbot': './walk',
    'brownian': './brownian',
    'blob': './blob',
    'punter2': './punter2',
    'random': '"python3 ./src/pybots/ai_random.py"',
    'trollbot': '"python3 ./src/pybots/trollbot.py"',
    'robbrown': '"python3 ./src/pybots/ai_olrobbrown.py"',
}


# Must call `docker build -t boxes .` first maybe
# docker network create icfp_local

def run(exe, port, host):
    if exe in BOT_INDEX:
        exe = BOT_INDEX[exe]

    cmd = 'docker run {volumes} {net} --rm boxes python3 {runner} {exe} {port} --server {host}'\
        .format(volumes=VOLS, net=NET, runner=RUNNER, exe=exe, port=port, host=host)

    # cmd = 'docker run --net=host --rm boxes ping {host}'.format(host=host)
    print(cmd)
    return subprocess.Popen(cmd, stdout=subprocess.PIPE, stdin=subprocess.PIPE)


def run_server(map, port, punters, host):
    exe = 'server'
    cmd = 'docker run {volumes} {net} --name {server} --expose {port} --rm boxes server -map {basedir}maps/{map} -port {port} -punters {punters} -runonce' \
        .format(
            volumes=VOLS,
            net=NET, map=map, exe=exe, port=port, basedir=BASEDIR, punters=punters,
            server=host
    )
    print(cmd)
    return subprocess.Popen(cmd.split(' '))  #, stdout=subprocess.PIPE, stdin=subprocess.PIPE)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='LET THEM FIGHT')
    parser.add_argument('--port', action="store", default=9000, type=int, help='')
    parser.add_argument('--map', action="store", default='sample.json', type=str, help='')
    parser.add_argument('--bot', action="append", type=str, help='')
    results = parser.parse_args()

    port = results.port
    map = results.map
    bots = results.bot
    host = SERVER+str(port)

    server_p = run_server(map, port, len(bots), host)
    sleep(1.0)
    bot_proc = []
    for bot in bots:
        p = run(exe=bot, port=port, host=host)
        bot_proc.append(p)

    server_p.communicate()
