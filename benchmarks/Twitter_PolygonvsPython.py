#Python 3.11
#Using https://github.com/JewishLewish/PolygonDB/blob/main/databases/database.json
import ast
from websocket import create_connection
import timeit
import json
from polywrapper import *

poly = PolyClient(connection_url="localhost:8080", dbname="twitter")

def Poly_Method():
    x = poly.read(location="")

def Python_Json_Methd():
    with open("databases/twitter.json", "r") as f:
        x= json.load(f)


def benchmark(func, num_runs=30):
    total_time = timeit.timeit(func, number=num_runs)
    print(f"\nAverage execution time over {num_runs} runs: {total_time / num_runs:.6f} seconds")

if __name__ == "__main__":
    benchmark(Poly_Method)
    benchmark(Python_Json_Methd)
