<div align="center">
	
![GitHub all releases](https://img.shields.io/github/downloads/JewishLewish/PolygonDB/total?color=63C9A4&style=for-the-badge)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/Jewishlewish/PolygonDB?color=63C9A4&style=for-the-badge)
![GitHub commit activity](https://img.shields.io/github/commit-activity/w/JewishLewish/PolygonDB?color=63C9A4&style=for-the-badge)
	
</div>

<div align="center"><h1>Polygon</h1></div>
<div align="center"><h4>Database System Designed to be Fast, Performant and Minimal</h4></div>
<hr>

[![Frame 2](https://user-images.githubusercontent.com/65754609/215379958-d8f02d22-fec4-4200-85c1-0177a62e661d.png)](https://discord.gg/heWJfMSMTm)

## 📖 [Wiki](https://github.com/JewishLewish/PolygonDB/wiki)

## ⚡️ Quickstart
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
```

## 🎯 Features
* Low Memory Usage
* Developer-Friendly
* Compatible with any lang (C to Python)
* Easy-to-Setup
* Customizable Password Security
* Takes Advantage of Synchronization
* Multi-thread safe

## 💡 Companies Who Use it

<div style="display: flex; justify-content: center;">
		<a href="https://discord.gg/muXKEkbRwp">

<img src="https://discordapp.com/api/guilds/692451473698586704/widget.png?style=banner2" alt="Discord Banner 2"/>
<img src="https://discordapp.com/api/guilds/879344703689064499/widget.png?style=banner2" alt="Discord Banner 2"/>
		</a>
	<a href="https://discord.gg/MHEAwNjKb2"><img src="https://discordapp.com/api/guilds/1024761808407498893/widget.png?style=banner2" alt="Discord Banner 2"/></a>
	<img src="https://discordapp.com/api/guilds/1046141941387116565/widget.png?style=banner2" alt="Discord Banner 2"/>
	<img src="https://discordapp.com/api/guilds/1076152760719900732/widget.png?style=banner2" alt="Discord Banner 2"/>
    <img src="https://discordapp.com/api/guilds/1067868449826685060/widget.png?style=banner2" alt="Discord Banner 2"/>
</div>

## 👀 Community Projects
| Name & Link | Description | Type |
|---------------|---------------------------------------------------| ------- |
| [PloyconJS](https://github.com/NekaouMike/PolyConJS) | NodeJS Package for Polygon | Package |
| [PolyDash](https://github.com/NekaouMike/PolyDash) | Polygon | Utility| 

## notice
If you wish for your company / module / Utility to be placed here make a request on our discord server.
