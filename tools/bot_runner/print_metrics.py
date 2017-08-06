#!/usr/bin/env python3

import argparse
from os import listdir
from os.path import isfile, join

from adapter import get_metrics


def main():
    parser = argparse.ArgumentParser(description='ICFP 2017 Metrics Generator')

    parser.add_argument('dir', action="store", type=str, help='directory of playlogs')
    parser.add_argument('--filter_bot', action="store", type=str, help='directory of playlogs')

    results = parser.parse_args()

    playlogs = [join(results.dir, f) for f in listdir(results.dir) if isfile(join(results.dir, f))]
    for log in playlogs:
        try:
            stats = get_metrics(log)
            if results.filter_bot and stats[2] != results.filter_bot:
                continue
            print('{}/{} \t{} \t{log} {}'.format(log=log, *stats))
        except:
            pass
            # print("couldn't parse:", log)


if __name__ == '__main__':
    main()
