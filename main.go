package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"html/template"
	"log"
	"net/http"
	"time"
)

var addr = flag.String("addr", "192.168.167.11:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options
var connectionCounter int = 0

var marketData = make(chan []byte)

var clientConnectedChannel = make(chan *websocket.Conn)
var clientErrorChannel = make(chan string)

func websocketClient() {
	log.Println("test")
}

func generateMarketData() {

	klaine, _, err := websocket.DefaultDialer.Dial("wss://stream.binance.com:9443/stream?streams=btcusdt@kline_1m.b10", nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
    // kline link wss://stream.binance.com:9443/stream?streams=btcusdt@kline_1m.b10
    
    //depth, _, err := websocket.DefaultDialer.Dial("wss://stream.binance.com:9443/stream?streams=!miniTicker@arr@3000ms/btcusdt@depth.b10/btcusdt@aggTrade.b10", nil)
	//if err != nil {
	//	log.Fatal("dial:", err)
	//}

	for {

		_, message, err := klaine.ReadMessage()
		if err != nil {
			log.Fatal("websocket read error:", err)
		}
		marketData <- message
		//marketData <- `{"stream":"btcusdt@kline_1m","data":{"e":"kline","E":1540559109164,"s":"BTCUSDT","k":{"t":1540559100000,"T":1540559159999,"s":"BTCUSDT","i":"1m","f":76335176,"L":76335179,"o":"6540.73000000","c":"6540.99000000","h":"6540.99000000","l":"6540.72000000","v":"0.09257900","n":4,"x":false,"q":"605.53653201","V":"0.00888200","Q":"58.09707318","B":"0"}}}`
		//time.Sleep(100 * time.Millisecond)
	}
}

type Client struct {
	id          string
	client      *websocket.Conn
	sendChannel chan []byte
}

func (cl *Client) Sender() {
	for {
		select {
		case msg := <-cl.sendChannel:
			err := cl.client.WriteMessage(1, []byte(msg))
			if err != nil {

				clientErrorChannel <- cl.id
				log.Printf("Error sending data %v", err)
				return
			}
		}
	}
}

func marketDataSenderRoutine() {
	clients := make(map[string]*Client)
	for {
		select {
		case c := <-clientConnectedChannel:

			uid, err := uuid.NewV4()
			//log.Printf("Client Connected %s", uid)
			if err != nil {
				//
				break
			}
			id := uid.String()
			client := &Client{id: id, client: c, sendChannel: make(chan []byte)}
			go client.Sender()
			clients[id] = client
			connectionCounter = connectionCounter + 1
		case id := <-clientErrorChannel:
			//log.Printf("Client Disconnected")
			client, ok := clients[id]
			if ok {
				client.client.Close()
				delete(clients, id)

				connectionCounter = connectionCounter - 1
			}
		case dataToSend := <-marketData:
			for _, client := range clients {
				client.sendChannel <- dataToSend
			}

		}

	}
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	clientConnectedChannel <- c

	//log.Print("upgrade:", err)
	//log.Printf("Connection number: %d", connectionCounter)

}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

func periodicPrints() {

	for {

		log.Printf("Connected users: %d", connectionCounter)
		time.Sleep(3 * time.Second)
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)
	go marketDataSenderRoutine()\
	go generateMarketData()

	go periodicPrints()
	log.Fatal(http.ListenAndServe(*addr, nil))

}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {

    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;

    var print = function(message) {
        var d = document.createElement("div");
        d.innerHTML = message;
        output.appendChild(d);
    };

    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };

    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };

    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };

});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))
