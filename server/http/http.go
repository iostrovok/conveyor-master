package http

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/iostrovok/conveyormaster/server/messager"
)

var upgrader = websocket.Upgrader{} // use default options

var indexTmpl *template.Template
var PWD string

func init() {

	PWD, _ = os.Getwd()

	t, err := template.ParseFiles(PWD + "/templates/index.html")
	if err != nil {
		log.Fatal(err)
	}
	indexTmpl = t
}

type Response struct {
	Error      string `json:"error"`
	OnlineIds  []int  `json:"online-ids"`
	OfflineIds []int  `json:"offline-ids"`
	IsDone     bool   `json:"is-done"`
	InProcess  bool   `json:"in-process"`
}

type Handler func(w http.ResponseWriter, req *http.Request)

// Start is an entry point for HTTP server
func Start(addr string, message messager.IMessage) error {

	r := mux.NewRouter()
	r.HandleFunc("/healthcare", healthCare)
	r.HandleFunc("/data", initHandlers(message))
	r.HandleFunc("/run", initRunHandler(message))

	//r.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("server/http/static/"))))
	//r.Handle("/", http.FileServer(http.Dir("static/")))

	//r.Handle("/static/", http.StripPrefix("/static/", fs))
	//r.Handle("/static/*", staticFunc(pwd+"/static/"))

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(PWD+"/static/"))))
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir(PWD+"/static/assets/"))))
	r.HandleFunc("/", home)

	fmt.Printf("PWD+/static/assets/ %s\n", PWD+"/static/assets/")
	fmt.Printf("....http started on %s\n", addr)

	return http.ListenAndServe(addr, r)
}

func Recover(msg string) func() {
	return func() {
		if rec := recover(); rec != nil {
			log.Printf("ERROR: %s.recover: %s", msg, string(debug.Stack()))
		}
	}
}

func initHandlers(message messager.IMessage) func(w http.ResponseWriter, req *http.Request) {

	return func(w http.ResponseWriter, req *http.Request) {

		defer Recover("initHandlers")

		err := jsonPrint(w, Response{IsDone: true})

		if err != nil {
			log.Printf("ERROR: %s", err.Error())
		}
	}
}

func healthCare(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	if err := jsonPrint(w, map[string]bool{"ok": true}); err != nil {
		log.Printf("ERROR: %s in healthCheck", err.Error())
	}
}

func jsonPrint(w http.ResponseWriter, data interface{}) error {
	b, err := json.Marshal(data)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	return err
}

func readMessage(conn *websocket.Conn) chan *messager.HttpMessage {

	ch := make(chan *messager.HttpMessage, 10)
	go func() {
		for {
			data := &messager.HttpMessage{}
			err := conn.ReadJSON(&data)
			if err != nil {
				log.Println("err read from http:", err)
				close(ch)
				return
			}
			ch <- data
		}
	}()

	return ch
}

func initRunHandler(ms messager.IMessage) func(w http.ResponseWriter, req *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		client := ms.HttpClient()
		defer ms.DeleteClient(client)

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer conn.Close()
		chIn := readMessage(conn)

		r.Context()

		ctx, cancel := context.WithCancel(r.Context())

		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {

				case <-ctx.Done():
					return
				case request, ok := <-chIn:
					if !ok {
						log.Println("ok read from http:", ok)
						cancel()
						return
					}

					log.Printf("recv: %s", request)
					ms.AddHttpRequest(request)
				}
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case request, ok := <-client.ReadHttpRequest():
					if !ok {
						log.Println("ok read from grpc:", ok)
						break
					}
					log.Printf("recv: %+v", request)

					err = conn.WriteJSON(request)
					if err != nil {
						log.Println("write grpc:", err)
						cancel()
						return
					}
				}
			}
		}()

		wg.Wait()
	}
}

func home(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("r.Host: %s\n", req.Host)
	//homeTemplate.Execute(w, "ws://"+req.Host+"/run")
	indexTmpl.Execute(w, "ws://"+req.Host+"/run")
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
        d.textContent = message;
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
