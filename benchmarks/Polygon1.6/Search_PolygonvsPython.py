#Python 3.10.9
#Using https://github.com/JewishLewish/PolygonDB/blob/main/databases/Search_Benchmark/database.json
import json
from websocket import create_connection
import timeit

#Searching through database for certain ID

ws = create_connection("ws://localhost:25565/ws")


def Poly_Method():
    ws.send(json.dumps(
        {
            'password': 'B123', 
            'dbname': 'Search_Benchmark',
            'location' :'data',
            'action' : 'search',
            'value' : 'guid:147bd43a-338c-450b-a293-4999dba1f367'
        }
    ))
    ws_data = json.loads(ws.recv())

def Python_Method():
    ws.send(json.dumps(
        {
            'password': 'B123', 
            'dbname': 'Search_Benchmark',
            'location' :'data',
            'action' : 'retrieve'
        }
    ))
    data = json.loads(ws.recv())
    # Create an empty list to store the male people
    males = []

    # Iterate through each person in the data
    for index, person in enumerate(data):
        # Check if the person's gender is male
        if person['guid'] == "147bd43a-338c-450b-a293-4999dba1f367":
            return


def benchmark(func, num_runs=90):
    total_time = timeit.timeit(func, number=num_runs)
    print(f"\nAverage execution time over 90 runs: {total_time / num_runs:.6f} seconds")

benchmark(Poly_Method)
benchmark(Python_Method)
