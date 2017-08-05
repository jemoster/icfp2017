#!/usr/bin/env python3

import subprocess
import argparse
import random
import os
from time import time

server = "punter.inf.ed.ac.uk"

bot_list = [
    "python -u src/pybots/ai_olrobbrown.py",
    "python -u src/pybots/ai_random.py",
    # "go run src/cmd/play/main.go",
]

def main():
    parser = argparse.ArgumentParser(description='ICFP 2017 Online Adapter')

    parser.add_argument('bot_count', action="store", type=int, help='how many bots total')
    parser.add_argument('port', action="store", type=int, help='Port of the competition server '
                        'see http://punter.inf.ed.ac.uk/status.html for details')
    parser.add_argument('--server', action="store", default="punter.inf.ed.ac.uk",
                        help='The game server to connect to defaults to "punter.inf.ed.ac.uk"')
    results = parser.parse_args()

    procs = []
    run_time = time()
    for num in range(results.bot_count):
        bot = random.choice(bot_list)
        basename = '{}_{}_{}_.1.txt'.format(results.port, run_time, num)
        filename = os.path.join('data', 'multibot', basename)
        cmd = " ".join(["python", "-u", "tools/bot_runner/online_adapter.py", '"{}"'.format(bot), str(results.port), '--record', filename])
        print('calling: ', cmd)
        procs.append(subprocess.Popen(cmd, shell=True))
    for proc in procs:
        proc.communicate()

if __name__ == "__main__":
    main()
