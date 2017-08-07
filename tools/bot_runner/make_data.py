#!/usr/bin/env python3

from online_adapter import OfflineAdapter
import random
import time
import argparse
import os

server = "punter.inf.ed.ac.uk"
ports = [
    (9003, 9005),
]


def pick_port():
    p_range = random.choice(ports)
    return random.choice(range(p_range[0], p_range[1]-1))


def main():
    parser = argparse.ArgumentParser(description='ICFP 2017 DATA CANNON')
    parser.add_argument('exe', action="store", help='The executable to evaluate')
    parser.add_argument('record', action="store", help='directory to save playlog to')
    results = parser.parse_args()

    while True:
        try:
            os.makedirs(results.record, exist_ok=True)

            port = pick_port()
            print('connecting to {}:{}'.format(server, port))

            log_name = os.path.join(results.record, '{}_{}'.format(port, time.strftime("%Y%m%d-%H%M%S")))
            print('logging to', log_name)

            adapter = OfflineAdapter(server, port, results.exe, log_name)
            scores = adapter.run()
            for player in scores:
                player_name = 'punter:' if player['punter'] != adapter.punter_id else "me:    "
                print('{} {punter}, score: {score}'.format(player_name, **player))
        except Exception as e:
            print("lol no we ain't stopping now!", e)

if __name__ == "__main__":
    main()
