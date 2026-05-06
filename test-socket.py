import asyncio
import websockets
import json

token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkcml2ZXJfaWQiOm51bGwsImVtYWlsIjoiZ2FiZ29sQGdtYWlsLmNvbSIsImV4cCI6MTc3ODA4ODU1MiwiaWQiOiI4NjVlNjc2Ni04MzRkLTRmYjQtYjUzNy05OGY3YTYyZjRmZTEiLCJuYW1lIjoiZ2FiZ29sIn0.CjElwfTFSBG9k_ESiPHiEh1OebUu43LMqH0je6kur70"

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