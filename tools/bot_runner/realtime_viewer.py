import asyncio
import queue

import websockets


async def realtime_handler(websocket, path, msg_queue, stop):
    print(path)
    config = await websocket.recv()

    while True:
        try:
            datagram = msg_queue.get(timeout=0.1)
        except queue.Empty:
            continue
        if datagram == 'stop':
            break

        await websocket.send(
            'data: {}'.format(datagram)
        )
    stop.done()


def realtime_viewer(queue=None):
    stop = asyncio.Future()
    start_server = websockets.serve(lambda x, y: realtime_handler(x, y, queue, stop), 'localhost', 5000)
    asyncio.get_event_loop().run_until_complete(start_server)
    asyncio.get_event_loop().run_until_complete(stop)