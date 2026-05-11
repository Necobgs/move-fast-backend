import asyncio
import websockets
import json

# MOTORISTA
token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkcml2ZXJfaWQiOiI4YzU5Yzg5Yi1iMzQ5LTQwYWYtOGEwYy05ODE3ZTNjNDNmMTEiLCJlbWFpbCI6Im1vdG9yaXN0YUBnbWFpbC5jb20iLCJleHAiOjE3NzgzODc5MjgsImlkIjoiNzljZjQ4MTktZWJiZS00Y2NmLTlkYzEtYmIyODFiMzM4OGNlIiwiaWRlbnRpZmllciI6ImRyaXZlcjo4YzU5Yzg5Yi1iMzQ5LTQwYWYtOGEwYy05ODE3ZTNjNDNmMTEiLCJuYW1lIjoiTW90b3Jpc3RhIn0.dtPBn5Ad-NKoTVgXUzkVYSbaW0vpvoH8eKrqWRwPRlY"

async def test():
    uri = "ws://26.24.78.160:8080/ws?token=" + token
    
    async with websockets.connect(uri) as websocket:
        msg = {
            "event": "update_location_driver",
            "data": {
                "lat": -28.9352,
                "lng": -49.4910,
            }
        }
        # msg2 = {
        #     "event": "requested_ride",
        #     "data": {
        #         "ride_id":"baabd4fd-dc61-4769-b2d2-6e23d66d1790",
        #         "accepted":True
        #     }
        # }
        await websocket.send(json.dumps(msg))
        response = await websocket.recv()
        print("Resposta:", response)
        input('...')
        # await websocket.send(json.dumps(msg2))
        # response = await websocket.recv()
        # print("Resposta2:", response)
        # input('...')



asyncio.run(test())