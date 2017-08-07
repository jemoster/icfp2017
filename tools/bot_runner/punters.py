#!/usr/bin/env python3

import argparse
from server_status import read_status
from collections import defaultdict


def list_players(status):
    players = defaultdict(
        lambda: {}
    )
    for game in status.values():
        for punter in game['punters']:
            players[punter][game['port']] = game
    return players


def filter_games(punters, waiting_limit=None, map_limit=None):
    filtered = {}
    for player in punters:
        games = punters[player]

        if waiting_limit:
            games = {k: g for k, g in games.items() if g['total_punters'] - len(g['punters']) < waiting_limit}

        if map_limit:
            games = {k: g for k, g in games.items() if g['map_name'] in map_limit}

        if games:
            filtered[player] = games

    return filtered


def main():
    parser = argparse.ArgumentParser(description='ICFP 2017 Online Adapter')
    parser.add_argument('--waiting', action="store", type=int, help='')
    parser.add_argument('--map', action="append", type=str, help='')
    results = parser.parse_args()

    punters = filter_games(
        list_players(
            read_status()
        ),
        results.waiting,
        results.map
    )

    for num, punter in enumerate(sorted(punters)):
        games = punters[punter]
        print('{:<3} {:<45} {:<4} : {}'.format(num, punter, len(games), games))


if __name__ == "__main__":
    main()
