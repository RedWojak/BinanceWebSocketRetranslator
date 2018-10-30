const WebSocket = require('ws');  // npm install ws
 
const u1 = 'wss://stream.binance.com:9443/stream?streams=!miniTicker@arr@3000ms/btcusdt@depth.b10/btcusdt@aggTrade.b10'
const u2 = 'wss://stream.binance.com:9443/ws/gMxuXiZIWS8lBlHxc8Qasl7ohZbLatZTQtekDh7iDQ1aLqJYwsFnMqQj6sgh'
const u3 = 'wss://stream.binance.com:9443/stream?streams=btcusdt@kline_1m.b10'
const u4 = 'ws://192.168.167.7:8881/ws/'
const u5 = 'ws://192.168.167.11:8080/echo'
let counter = 0
let timerSet = false
const startTime = new Date().getTime()/1000

const doWs = (id) => {
    let ws = new WebSocket(u5);
    ws.on('open', function open() {
        // ws.send('Connection #' + id + ' has been established!')
        console.log('opened ', id)
    });

    ws.on('error', (err) => {
        console.log(err)
    })
    
    ws.on('message', function incoming(data) {
        counter++
        // const d = JSON.parse(data)
        const d = data
        // ws.send('Connection #' + id + ' sends its regards!')
        if (id === 0) {
            console.log(counter, d, d.length, Math.floor(new Date().getTime()/1000 - startTime));
        }
      });
}

 
for (let i = 0; i < 1; i++) {
    doWs(i)
}
