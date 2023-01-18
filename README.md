# PolygonDB
Polygon Database is a portable Database System that allows users to repurpose unused and small servers into databases


# Usage
Adjust databases/name-of-server/config.json
```json
{
    "path":"database.json",
    "key":"Better_Password"
}
```

Database.json example
```json
{
	"rows": [
		{
			"age": 5,
			"name": "A"
		},
		{
			"age": 20,
			"name": "B"
		},
		{
			"age": 30,
			"name": "C"
		}
	]
}
```
Python code for accessing the server
```python
import json
from websocket import create_connection


ws = create_connection("ws://localhost:8000/ws")

ws.send(json.dumps(
    {
        'password': 'Secret_Password', 
        'dbname': 'CatoDB',
        'location' :'rows.0.name',
        'action' : 'retrieve'
    }
))
print(json.loads(ws.recv()))  # "A"

ws.send(json.dumps(
    {
        'password': 'Secret_Password', 
        'dbname': 'CatoDB',
        'location' :'rows.0.age',
        'action' : 'record',
        'value':'5'
    }
))
print(json.loads(ws.recv())) # {Status: Success}

ws.send(json.dumps(
    {
        'password': 'Secret_Password', 
        'dbname': 'CatoDB',
        'location' :'rows',
        'action' : 'search',
        'value':'age:30'
    }
))
print(json.loads(ws.recv())) # {'age': 30, 'name': 'C'}```
```

# Companies that uses PolygonDB
![Discord Banner 2](https://discordapp.com/api/guilds/1024761808407498893/widget.png?style=banner2)

Sidenote, if you want your server to be on here then open an issue. 
