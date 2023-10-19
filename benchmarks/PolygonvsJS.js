//Node.js v18.15.0
//Using https://github.com/JewishLewish/PolygonDB/blob/main/databases/database.json
const fs = require('fs');
const Benchmarkify = require("benchmarkify");
const benchmark = new Benchmarkify("Polygon vs NodeJS via File Reading").printHeader();
const bench1 = benchmark.createSuite("Polygon vs NodeJS");


const WebSocket = require('ws');
const ws = new WebSocket('ws://localhost:8080/ws');
ws.setMaxListeners(0);

const sendData = (e) => {
    ws.on('open', () => {
      ws.send(JSON.stringify(e));
    });
    };


function polyMethod() {
    const data1 = {
    dbname: 'Search_Benchmark',
    location: 'database',
    action: 'read',
    value: '_'
  };
  sendData(data1);
}

function nodeMethod() {
    try {
        const data = JSON.parse(fs.readFileSync('databases/database.json', 'utf8')).data;
        return data;
      } catch (error) {
        console.error('Error reading data from JSON file:', error);
        return null;
      }
}

ws.on('message', message => {
    const response = JSON.parse(message);
  });


bench1.add("Using PolygonDB", () => {
  for (let i = 0; i < 90; i++) {
    polyMethod();
  }
});

bench1.ref("Using NodeJS", () => {
  for (let i = 0; i < 90; i++) {
    nodeMethod();
  }
});



bench1.run();