#!/usr/bin/env python3

import json
import utils
import argparse
from os import listdir
from os.path import isfile, join


def parse(line):
    return json.loads(line[2:])


def get_rank(punter_id, scores):
    for rank, score in enumerate(utils.ranked(scores)):
        if punter_id == score['punter']:
            return rank+1
    return None


def get_metrics(filename):
    with open(filename) as f:
        name = parse(f.readline())['me']
        f.readline()
        punter_id = parse(f.readline())['punter']

        # Skip to the end of the file
        for line in f:
            pass

        score = parse(line)
        if 'stop' not in score:
            return

        scores = score['stop']['scores']
        rank = get_rank(punter_id, scores)
        return rank, len(scores), name


def main():
    parser = argparse.ArgumentParser(description='ICFP 2017 Metrics Generator')

    parser.add_argument('dir', action="store", type=str, help='directory of playlogs')
    results = parser.parse_args()

    playlogs = [join(results.dir, f) for f in listdir(results.dir) if isfile(join(results.dir, f))]
    for log in playlogs:
        try:
            stats = get_metrics(log)
            print('{}/{} \t{} \t{log}'.format(log=log, *stats))
        except:
            pass
            # print("couldn't parse:", log)



if __name__ == '__main__':
    main()
