from online_adapter import OfflineAdapter
import random
from time import time
import argparse

server = "punter.inf.ed.ac.uk"
ports = [
    (9003, 9010),
    (9016, 9019),
    (9026, 9030),
    (9036, 9040),
    (9046, 9050),
]


def pick_port():
    p_range = random.choice(ports)
    return random.choice(range(p_range[0], p_range[1]-1))


def main():
    parser = argparse.ArgumentParser(description='ICFP 2017 DATA CANNON')
    parser.add_argument('exe', action="store", help='The executable to evaluate')
    results = parser.parse_args()

    while True:
        try:
            port = pick_port()
            print('connecting to {}:{}'.format(server, port))
            log_name = 'data/{}_{}.0'.format(port, time())
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