import subprocess
import argparse

# Note that this requires you have the directory checked out to /tmp/maps
BASEDIR = '/src/github.com/jemoster/icfp2017/'
HOSTDIR = 'c:/Users/Joe/Documents/src/icfp2017/'

MAPS = '-v {}maps:{}maps'.format(HOSTDIR, BASEDIR)
PLAYLOG = '-v {}playlogs:{}data'.format(HOSTDIR, BASEDIR)

VOLS = '{} {}'.format(MAPS, PLAYLOG)
RUNNER = './tools/bot_runner/online_adapter.py'
PLAYER = './tools/bot_runner/make_player_data.py'
IDLERUN = './tools/bot_runner/idle_runner.py'
IDLENAME = 'idle-tmp'


BOT_INDEX = {
    'simpleton': './simpleton',
    'walkbot': './walk',
    'brownian': './brownian',
    'blob': './blob',
    'punter2': './punter2',
    'random': './src/pybots/ai_random.py',
    'trollbot': './src/pybots/trollbot.py',
    'robbrown': './src/pybots/ai_olrobbrown.py',
}


# Must call `docker build -t boxes .` first


def run(exe, port, host):
    if exe in BOT_INDEX:
        exe = BOT_INDEX[exe]

    cmd = 'docker run {volumes} --net=host --rm boxes python3 {runner} {exe} {port} --server {host}'\
        .format(volumes=VOLS, runner=RUNNER, exe=exe, port=port, host=host)
    print(cmd)
    return subprocess.Popen(cmd, stdout=subprocess.PIPE, stdin=subprocess.PIPE)


def run_server(map, port, punters):
    exe = 'server'
    cmd = 'docker run {volumes} --expose {port} --rm boxes server -map {basedir}maps/{map} -port {port} -punters {punters}' \
        .format(volumes=VOLS, map=map, exe=exe, port=port, basedir=BASEDIR, punters=punters)
    print(cmd)
    return subprocess.Popen(cmd.split(' ')) #, stdout=subprocess.PIPE, stdin=subprocess.PIPE)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='LET THEM FIGHT')
    parser.add_argument('--port', action="store", default=9000, type=int, help='')
    parser.add_argument('--map', action="store", default='sample.json', type=str, help='')
    parser.add_argument('--bot', action="append", type=str, help='')
    results = parser.parse_args()

    port = results.port
    map = results.map
    host = '127.0.0.1'

    server_p = run_server(map, port, len(results.bot))
    bot_proc = []
    for bot in results.bot:
        p = run(exe=bot, port=port, host=host)
        bot_proc.append(p)

    server_p.communicate()
