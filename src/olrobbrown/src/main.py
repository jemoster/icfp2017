import sys
import json
import random

def setup(setup):
    p = setup['punter']

    possible_claims = []
    for river in setup['map']['rivers']:
        possible_claims.append((river['source'], river['target']))

    punter_id = setup['punter']

    # Start with first mine
    available_mines = setup['map']['mines']
    start_mine = random.choice(available_mines)
    available_mines.remove(start_mine)
    prev_sites = [start_mine]
    return ({'ready': p,
        'state': {
            'possible_claims': possible_claims,
            'punter_id': punter_id,
             'available_mines': available_mines,
            'prev_sites':prev_sites}}
    )

def choose_first_connected(claims, prev_sites, last_end):
    for claim in claims:
        if claim[0] == last_end:
            if not claim[1] in prev_sites:
                return claim[1], claim
        if claim[1] == last_end:
            if not claim[0] in prev_sites:
                return claim[0], claim
    return None, None

def gameplay(msg):
    if 'stop' in msg:
        return {}

    if 'timeout' in msg:
        return {}

    possible_claims = msg['state']['possible_claims']
    punter_id = msg['state']['punter_id']
    available_mines = msg['state']['available_mines']
    prev_sites = msg['state']['prev_sites']

    for move in msg['move']['moves']:
        if 'pass' in move:
            continue
        move = move['claim']
        possible_claims.remove([move['source'], move['target']])

    last_idx = -1
    next_site, claim = choose_first_connected(possible_claims, prev_sites, prev_sites[last_idx])
    while claim is None:
        last_idx -= 1
        if -last_idx > len(prev_sites):
            break
        next_site, claim = choose_first_connected(possible_claims, prev_sites, prev_sites[last_idx])

    if claim is None and available_mines:
        next_mine = random.choice(available_mines)
        available_mines.remove(next_mine)
        next_site, claim = choose_first_connected(possible_claims, prev_sites, next_mine)

    if claim is None:
        claim = random.choice(possible_claims)
        next_site = claim[0]

    prev_sites.append(next_site)

    return (
        {"claim": {"punter": punter_id,
         'source': claim[0],
         'target': claim[1]},
         'state': {
             'possible_claims': possible_claims,
             'punter_id': punter_id,
             'available_mines': available_mines,
             'prev_sites': prev_sites}}
    )


def format_send(msg):
    serialized_msg = json.dumps(msg)
    return "{}:{}".format(len(serialized_msg), serialized_msg)


def read_structured(buffer):
    while ':' not in buffer:
        buffer += sys.stdin.read(1)

    buffer_size_txt = buffer.split(':', 1)[0]
    msg_size = int(buffer_size_txt)
    min_buffer_size = len(buffer_size_txt) + msg_size + 1

    while len(buffer) < min_buffer_size:
        buffer += sys.stdin.read(1)

    msg_txt = buffer[:min_buffer_size]
    buffer = buffer[min_buffer_size:]
    return json.loads(msg_txt.split(':', 1)[1]), buffer


def print_err(msg):
    sys.stderr.write("{}\n".format(msg))
    sys.stderr.flush()

if __name__ == '__main__':
    buffer = ''

    # Handshake
    handshake = {'me': 'Ol\' Rob Brown'}
    print(format_send(handshake))
    hand_in, buffer = read_structured(buffer)


    # Execute for state update
    msg_in, buffer = read_structured(buffer)
    if 'punter' in msg_in:
        msg = setup(msg_in)
    else:
        msg = gameplay(msg_in)

    print(format_send(msg))
