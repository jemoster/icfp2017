import asyncio
import websockets
import json

def parse(line):
    return json.loads(line[2:])

async def hello(websocket, path):
    print(path)
    config = await websocket.recv()
    file = config.split(':', 1)[1].rsplit(':', 1)[0]
    print('loading...', file)

    with open(file) as f:
        print('opened')
        for line in f:
            if line[:2] != '<<':
                continue

            msg = parse(line)
            if 'you' in msg:
                continue

            print('sending...')
            await websocket.send(
                'data: {}'.format(json.dumps(msg))
            )
            await asyncio.sleep(2)

start_server = websockets.serve(hello, 'localhost', 5000)

asyncio.get_event_loop().run_until_complete(start_server)
asyncio.get_event_loop().run_forever()