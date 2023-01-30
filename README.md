<div align="center"><h1>Polygon</h1></div>
<div align="center"><h4>Database designed to be Minimal</h4></div>
<hr>

![Frame 2](https://user-images.githubusercontent.com/65754609/215379958-d8f02d22-fec4-4200-85c1-0177a62e661d.png)

## Details about Project
https://github.com/JewishLewish/PolygonDB/wiki

## Usage
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
print(json.loads(ws.recv())) # {'Index': 2, 'Value': {'age': 30, 'name': 'C'}}
```

## Companies that uses PolygonDB

<div style="display: flex; justify-content: center;">
	<img src="https://discordapp.com/api/guilds/1024761808407498893/widget.png?style=banner2" alt="Discord Banner 2"/>
	<img src="https://discordapp.com/api/guilds/1046141941387116565/widget.png?style=banner2" alt="Discord Banner 2"/>
</div>

<img src="https://discordapp.com/api/guilds/692451473698586704/widget.png?style=banner2" alt="Discord Banner 2"/>

## Modules / Packages for Certain Langs
Javascript - https://github.com/NekaouMike/PloyConJS


If you wish for your company / module to be placed here then open a pull request
