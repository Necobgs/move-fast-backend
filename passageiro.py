import asyncio
import websockets
import json

# PASSAGEIRO
token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkcml2ZXJfaWQiOm51bGwsImVtYWlsIjoicGFzc2FnZWlyb0BnbWFpbC5jb20iLCJleHAiOjE3NzgzODc5NDMsImlkIjoiOTUyNmFmNGItNjk3Mi00MDBiLTk4NjUtMTU4MDQzMWU3YWMyIiwiaWRlbnRpZmllciI6InVzZXI6OTUyNmFmNGItNjk3Mi00MDBiLTk4NjUtMTU4MDQzMWU3YWMyIiwibmFtZSI6IlBhc3NhZ2Vpcm8ifQ.pybCYFmZFEM2WjmOhJIMCnKfSswFzt1o0cxbTd0rksQ"

async def test():
    uri = "ws://26.24.78.160:8080/ws?token=" + token
    
    async with websockets.connect(uri) as websocket:
        msg = {
            "event": "request_ride",
            "data": {
                "origin_location_lat": -28.9352,
                "origin_location_lng": -49.4910,
                "origin_location_address": "Rua XV de Novembro, 123 - Centro, Araranguá - SC",

                "destination_location_lat": -28.6775,
                "destination_location_lng": -49.3697,
                "destination_location_address": "Av. Centenário, 4500 - Centro, Criciúma - SC"
            }
        }
        await websocket.send(json.dumps(msg))
        response = await websocket.recv()
        print("Resposta:", response)
        input('...')


asyncio.run(test())