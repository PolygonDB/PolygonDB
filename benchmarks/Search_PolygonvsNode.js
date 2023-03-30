//Node.js v18.15.0
//Using https://github.com/JewishLewish/PolygonDB/blob/main/databases/Search_Benchmark/database.json
const Benchmarkify = require("benchmarkify");
const benchmark = new Benchmarkify("Polygon vs NodeJS via Index").printHeader();
const bench1 = benchmark.createSuite("Polygon vs NodeJS");



const WebSocket = require('ws');
const ws = new WebSocket('ws://localhost:25565/ws');

const sendData = (e) => {
    ws.on('open', () => {
      ws.send(JSON.stringify(e));
    });
    };


function polyMethod() {
    const data1 = {
    password: 'B123',
    dbname: 'Search_Benchmark',
    location: 'data',
    action: 'search',
    value: 'guid:147bd43a-338c-450b-a293-4999dba1f367'
  };
  sendData(data1);
}

function nodeMethod() {
  const data2 = {
    'password': 'B123',
    'dbname': 'Search_Benchmark',
    'location': 'data',
    'action': 'retrieve'
  };

  sendData(data2);

  ws.on('message', message => {
    const response = JSON.parse(message);

    // Create an empty list to store the male people
    var males = [];

    var person = response[index];

    // Iterate through each person in the response data
    for (let index = 0; index < response.length; index++) {

      // Check if the person's gender is male
      if (person["guid"] == "147bd43a-338c-450b-a293-4999dba1f367") {
        break
      }
    }
  
  });
}

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