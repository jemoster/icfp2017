#!/usr/bin/env python3

import subprocess
import argparse
import random
import os
from time import time

server = "punter.inf.ed.ac.uk"

known_bots = {
    'brown': "python -u src/pybots/ai_olrobbrown.py",
    'random': "python -u src/pybots/ai_random.py",
    # "go run src/bots/prattmic/walk/main.go",
    # "go run src/bots/unremarkable/simpleton/main.go",
}

def main():
    parser = argparse.ArgumentParser(description='ICFP 2017 Online Adapter')

    parser.add_argument('bot_count', action="store", type=int, help='how many bots total')
    parser.add_argument('port', action="store", type=int, help='Port of the competition server '
                        'see http://punter.inf.ed.ac.uk/status.html for details')
    parser.add_argument('--bot', action='append', default=[], help='Add bots that you want tested')
    parser.add_argument('--default', action='store', default=None, help='default bot to use')
    parser.add_argument('--server', action="store", default="punter.inf.ed.ac.uk",
                        help='The game server to connect to defaults to "punter.inf.ed.ac.uk"')
    results = parser.parse_args()

    bot_list = results.bot

    procs = []
    run_time = time()
    for num in range(results.bot_count):
        if bot_list:
            bot = bot_list.pop()
        elif results.default:
            bot = results.default
        else:
            bot = random.choice(list(known_bots.keys()))

        if bot in known_bots:
            bot = known_bots[bot]

        basename = '{}_{}_{}_'.format(results.port, run_time, num)
        filename = os.path.join('data', 'multibot', basename)
        cmd = " ".join(["python", "-u", "tools/bot_runner/online_adapter.py", '"{}"'.format(bot), str(results.port), '--record', filename])
        print('calling: ', cmd)
        procs.append(subprocess.Popen(cmd, shell=True))
    for proc in procs:
        proc.communicate()

if __name__ == "__main__":
    main()
