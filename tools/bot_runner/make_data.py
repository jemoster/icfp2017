#!/usr/bin/env python3

from online_adapter import OfflineAdapter
import random
from time import time
import argparse
import os

server = "punter.inf.ed.ac.uk"
ports = [
    (9003, 9010),
    (9016, 9019),
    (9026, 9030),
    (9036, 9040),
    (9046, 9050),
    (9056, 9060),
    (9066, 9070),
    (9076, 9080),
    (9086, 9090),
    (9096, 9100),
    (9106, 9110),
    (9116, 9120),
    (9126, 9130),
    (9136, 9140),
    (9146, 9150),
    (9156, 9160),
    (9166, 9170),
    (9176, 9180),
    (9186, 9190),
    (9196, 9200),
    (9206, 9210),
    (9216, 9220),
    (9226, 9230),
    (9236, 9240),
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
            if not os.path.exists(results.record):
                os.makedirs(results.record)

            port = pick_port()
            print('connecting to {}:{}'.format(server, port))

            log_name = os.path.join(results.record, '{}_{}'.format(port, time()))
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