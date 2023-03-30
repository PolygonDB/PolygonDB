#Python 3.10.9
#Using https://github.com/JewishLewish/PolygonDB/blob/main/databases/Search_Benchmark/database.json
import json
from websocket import create_connection
import time

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


def benchmark(func, *args, **kwargs):
    total_time = 0
    num_runs = 90
    
    for i in range(num_runs):
        start_time = time.time()
        result = func(*args, **kwargs)
        total_time += time.time() - start_time
        #print(f"Run {i+1}: Function {func.__name__} took {elapsed_time:.6f} seconds to execute.")
    
    avg_time = total_time / num_runs
    print(f"\nAverage execution time over {num_runs} runs: {avg_time:.6f} seconds")
    return avg_time

Poly_Result = benchmark(Poly_Method)

Py_Result = benchmark(Python_Method)