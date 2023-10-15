#Python 3.10.9
#Using https://github.com/JewishLewish/PolygonDB/blob/main/databases/Search_Benchmark/database.json
from websocket import create_connection
import timeit
import json
from polywrapper import *

poly = PolyClient(connection_url="localhost:8080", dbname="database")

def Poly_Method():
    x = poly.read(location="data")

def Python_Json_Methd():
    with open("databases/database.json", "r") as f:
        x= json.load(f)["data"]

def benchmark(func, num_runs=90):
    total_time = timeit.timeit(func, number=num_runs)
    print(f"\nAverage execution time over 90 runs: {total_time / num_runs:.6f} seconds")

benchmark(Poly_Method)
benchmark(Python_Json_Methd)