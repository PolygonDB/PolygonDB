<div align="center">
	
![GitHub all releases](https://img.shields.io/github/downloads/JewishLewish/PolygonDB/total?color=63C9A4&style=for-the-badge)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/Jewishlewish/PolygonDB?color=63C9A4&style=for-the-badge)
![GitHub commit activity](https://img.shields.io/github/commit-activity/w/JewishLewish/PolygonDB?color=63C9A4&style=for-the-badge)
	
</div>

<div align="center"><h1>Polygon</h1></div>
<div align="center"><h4>Database System Designed to be Fast, Performant and Minimal</h4></div>
<hr>

![Frame 2](https://user-images.githubusercontent.com/65754609/215379958-d8f02d22-fec4-4200-85c1-0177a62e661d.png)

## Wiki
[HERE](https://github.com/JewishLewish/PolygonDB/wiki)
## Discord/Support
[HERE](https://discord.gg/heWJfMSMTm)
## Usage
(more examples in wiki)
Config.json Example
```json
{
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
Javascript Code via [PolyConJS](https://github.com/NekaouMike/PolyConJS)
```js
const ploycon = require('ployconjs');

const db = new ploycon("localhost:8000","Secret_Password");

main();
async function main() {
    await db.open().catch(err => console.log(err))
    console.log(await db.retrieve('CatoDB',0,"name"))
    console.log(db.record('CatoDB',0,"age",5))
    console.log(await db.search('CatoDB','age',30))
    console.log(await db.getschema('CatoDB'))
    console.log(await db.append('CatoDB',{
        "age": 20,
        "name": "Example"
    }))
        db.close()
}
```
## Modules / Packages for Certain Langs
Javascript - [PolyconJS](https://github.com/NekaouMike/PolyConJS)

## Utilities 
Dashboard - [PolyDash](https://github.com/NekaouMike/PolyDash)

## Companies that uses PolygonDB 

<div style="display: flex; justify-content: center;">
		<a href="https://discord.gg/muXKEkbRwp">

<img src="https://discordapp.com/api/guilds/692451473698586704/widget.png?style=banner2" alt="Discord Banner 2"/>
<img src="https://discordapp.com/api/guilds/879344703689064499/widget.png?style=banner2" alt="Discord Banner 2"/>
		</a>
	<a href="https://discord.gg/MHEAwNjKb2"><img src="https://discordapp.com/api/guilds/1024761808407498893/widget.png?style=banner2" alt="Discord Banner 2"/></a>
	<img src="https://discordapp.com/api/guilds/1046141941387116565/widget.png?style=banner2" alt="Discord Banner 2"/>
	<img src="https://discordapp.com/api/guilds/1076152760719900732/widget.png?style=banner2" alt="Discord Banner 2"/>
</div>

## notice
If you wish for your company / module / Utility to be placed here make a request on our discord server.
