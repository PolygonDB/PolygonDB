//Node.js v18.15.0
//Using https://github.com/JewishLewish/PolygonDB/blob/main/databases/Search_Benchmark/database.json
const WebSocket = require('ws');
const ws = new WebSocket('ws://localhost:25565/ws');

function polyMethod() {
  const data = {
    password: 'B123',
    dbname: 'Search_Benchmark',
    location: 'data',
    action: 'index',
    value: 'gender:male'
  };

  const sendData = () => {
    if (ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(data));
    } else {
      setTimeout(sendData, 10);
    }
  };

  sendData();
  
  ws.on('message', message => {
    const response = JSON.parse(message);
  });
}

function nodeMethod() {
  const { spawn } = require('child_process');
  const pyProg = spawn('python', ['./python/benchmark.py']);

  pyProg.stdout.on('data', function(data) {
    console.log(`Python output: ${data}`);
  });
}

function benchmark(func) {
  const numRuns = 90;
  let totalTime = 0;

  for (let i = 0; i < numRuns; i++) {
    const startTime = new Date().getTime();
    func();
    const endTime = new Date().getTime();
    const elapsedTime = endTime - startTime;
    totalTime += elapsedTime;
    //console.log(`Run ${i + 1}: Function ${func.name} took ${elapsedTime} milliseconds to execute.`);
  }

  const averageTime = totalTime / numRuns;
  console.log(`\nAverage execution time over ${numRuns} runs: ${averageTime.toFixed(6)} milliseconds`);
  return averageTime;
}

const polyResult = benchmark(polyMethod);
const pyResult = benchmark(nodeMethod);
