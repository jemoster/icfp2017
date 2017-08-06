#!/usr/bin/env python3

import argparse
from adapter import OfflineAdapter


def main():
    parser = argparse.ArgumentParser(description='ICFP 2017 Online Adapter')

    parser.add_argument('exe', action="store", help='The executable to evaluate')
    parser.add_argument('port', action="store", type=int, help='Port of the competition server '
                        'see http://punter.inf.ed.ac.uk/status.html for details')
    parser.add_argument('--server', action="store", default="punter.inf.ed.ac.uk",
                        help='The game server to connect to defaults to "punter.inf.ed.ac.uk"')
    parser.add_argument('--record', action="store", default=None, help='filename to save playlog to')
    parser.add_argument('--header', action="store", type=str, default=None)
    results = parser.parse_args()

    adapter = OfflineAdapter(results.server, results.port, results.exe, results.record, results.header)
    scores = adapter.run()

    for player in scores:
        player_name = 'punter:' if player['punter'] != adapter.punter_id else "me:    "
        print('{} {punter}, score: {score}'.format(player_name, **player))

if __name__ == "__main__":
    main()
