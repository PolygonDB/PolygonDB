package main

/*
#include <stdio.h>
#include "sysadmin/sysadmin.c"
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/Jeffail/gabs/v2"

	"github.com/gorilla/websocket"
)

type config struct {
	Key string `json:"key"`
}

type settings struct {
	Addr  string `json:"addr"`
	Port  string `json:"port"`
	Sbool bool   `json:"allow_remote_connections_sysadmin"`
	Spass string `json:"system_remote_connection_password"`
}

// main
func main() {

	var set settings
	portgrab(&set)

	if set.Sbool == true {
		http.HandleFunc("/terminal", Terminal)
	}

	http.HandleFunc("/ws", datahandler)
	fmt.Print("Server started on -> "+set.Addr+":"+set.Port, "\n")
	go mainTerm()
	go processQueue(queue)
	http.ListenAndServe(set.Addr+":"+set.Port, nil)
}

func portgrab(set *settings) {
	file, _ := os.ReadFile("settings.json")
	json.Unmarshal(file, &set)
	file = nil
}

// data handler
var databases = &sync.Map{}
var upgrader = websocket.Upgrader{
	EnableCompression: true,
	ReadBufferSize:    0,
	WriteBufferSize:   0,
}

type wsMessage struct {
	ws  *websocket.Conn
	msg map[string]interface{}
}

var queue = make(chan wsMessage)

func datahandler(w http.ResponseWriter, r *http.Request) {

	ws, _ := upgrader.Upgrade(w, r, nil)
	defer ws.Close()

	for {
		//Reads input
		var msg map[string]interface{}
		ws.ReadJSON(&msg)
		if msg == nil {
			break
		}
		queue <- wsMessage{ws: ws, msg: msg}
		msg = nil
	}
}

func processQueue(queue chan wsMessage) {
	for {
		msg := <-queue
		process(&msg.msg, msg.ws)
		Nullify(&msg)
	}
}

// Processes the request
// Once request is done, it cleans up out-of-scope variables
func process(msg *map[string]interface{}, ws *websocket.Conn) {

	var confdata config
	var database gabs.Container

	dbfilename := (*msg)["dbname"].(string)
	er := cd(&dbfilename, &confdata, &database)
	if er != nil {
		ws.WriteJSON("{Error: " + er.Error() + ".}")
		return
	}
	if (*msg)["password"] != confdata.Key {
		ws.WriteJSON("{Error: Password Error.}")
		return
	}
	Nullify(&confdata)

	direct := (*msg)["location"].(string)
	action := (*msg)["action"].(string)

	if action == "retrieve" {
		output := retrieve(&direct, &database)
		ws.WriteJSON(&output)
	} else {
		value := []byte((*msg)["value"].(string))
		if action == "record" {
			output := record(&direct, &database, &value, &dbfilename)
			ws.WriteJSON("{Status: " + output + "}")
		} else if action == "search" {
			output := search(&direct, &database, &value)
			ws.WriteJSON(&output)
		} else if action == "append" {
			output := append(&direct, &database, &value, &dbfilename)
			ws.WriteJSON("{Status: " + output + "}")
		}
	}

	//When the request is done, it sets everything to either nil or nothing. Easier for GC.
	Nullify(&database)
	runtime.GC()

}

// Config and Database Getting
// Uses Concurrency to speed up this process and more precised error handling
func cd(location *string, jsonData *config, database *gabs.Container) error {
	if _, err := os.Stat("databases/" + *location); !os.IsNotExist(err) {
		var conferr error
		var dataerr error

		conf(&conferr, &*location, &*jsonData)

		if value, ok := databases.Load(&*location); *&ok {
			*database = parsedata(&value)
			value = nil
		} else {
			data(&dataerr, &*location, &*database)
		}

		if conferr != nil {
			return conferr
		}
		if dataerr != nil {
			return dataerr
		}
		return nil

	} else {
		return err
	}
}

// This gets the database file
func data(err *error, location *string, database *gabs.Container) {
	var file []byte
	file, *err = os.ReadFile("databases/" + *location + "/database.json")
	if *err != nil {
		go fmt.Println("Error reading file:", err)
	}

	// Unmarshal the JSON data into a variable
	var data interface{}
	*err = json.Unmarshal(*&file, &data)
	if *err != nil {
		go fmt.Println("Error unmarshalling Database JSON:", err)
	}
	*database = parsedata(&data)
	databases.Store(*location, *&data)
	data = nil
	file = nil
	//*done <- true
}

func conf(err *error, location *string, jsonData *config) {
	var file []byte
	file, *err = os.ReadFile("databases/" + *location + "/config.json")
	if *err != nil {
		go fmt.Println("Error reading file:", err)
	}
	// Unmarshal the JSON data for config
	*err = json.Unmarshal(*&file, &jsonData)
	if *err != nil {
		go fmt.Println("Error unmarshalling Config JSON:", err)
	}
	file = nil
}

// Types of Actions
func retrieve(direct *string, jsonParsed *gabs.Container) interface{} {
	if *direct == "" {
		return jsonParsed.String()
	} else {
		return jsonParsed.Path(*direct).String()
	}
}

func record(direct *string, jsonParsed *gabs.Container, value *[]byte, location *string) string {

	val, err := UnmarshalJSONValue(&*value)
	if err != nil {
		return "Failure. Value cannot be unmarshal to json."
	}

	_, err = jsonParsed.SetP(&val, *direct)
	if err != nil {
		return "Failure. Value cannot be placed into database."
	}

	syncupdate(*jsonParsed, *&location)

	return "Success"
}

func search(direct *string, jsonParsed *gabs.Container, value *[]byte) interface{} {
	parts := strings.Split(string(*value), ":")
	targ := []byte(parts[1])
	target, _ := UnmarshalJSONValue(&targ)
	targ = nil

	var output interface{}

	it := jsonParsed.Path(*direct).Children()
	for i, user := range it {
		if fmt.Sprint(user.Path(parts[0]).Data()) == fmt.Sprint(*&target) {
			output = map[string]interface{}{"Index": i, "Value": user.Data()}
			break
		}
	}

	return output
}

func append(direct *string, jsonParsed *gabs.Container, value *[]byte, location *string) string {

	val, err := UnmarshalJSONValue(&*value)
	if err != nil {
		return "Failure. Value cannot be unmarshal to json."
	}

	er := jsonParsed.ArrayAppendP(&val, *direct)
	if er != nil {
		return "Failure!"
	}

	syncupdate(*jsonParsed, *&location)

	return "Success"
}

// Unmarhsals the value into an appropriate json input
func UnmarshalJSONValue(data *[]byte) (interface{}, error) {
	var v interface{}
	var err error
	if len(*data) == 0 {
		return nil, fmt.Errorf("json data is empty")
	}
	switch (*data)[0] {
	case '"':
		if (*data)[len(*data)-1] != '"' {
			return nil, fmt.Errorf("json string is not properly formatted")
		}
		v = string((*data)[1 : len(*data)-1])
	case '{':
		if (*data)[len(*data)-1] != '}' {
			return nil, fmt.Errorf("json object is not properly formatted")
		}
		err = json.Unmarshal(*data, &v)
	case '[':
		if (*data)[len(*data)-1] != ']' {
			return nil, fmt.Errorf("json array is not properly formatted")
		}
		err = json.Unmarshal(*data, &v)
	default:
		i, e := strconv.Atoi(string(*data))
		if e != nil {
			v = string(*data)
			return v, err
		}
		v = i
	}
	return v, err
}

// parses database
func parsedata(database interface{}) gabs.Container {
	jsonData, _ := json.Marshal(&database)
	Nullify(&database)
	jsonParsed, _ := gabs.ParseJSON(*&jsonData)
	return *jsonParsed
}

// Nullify basically helps with the memory management when it comes to websockets
func Nullify(ptr interface{}) {
	val := reflect.ValueOf(*&ptr)
	if val.Kind() == reflect.Ptr {
		val.Elem().Set(reflect.Zero(val.Elem().Type()))
	}
}

// Sync Update
func syncupdate(jsonParsed gabs.Container, location *string) {
	jsonData, _ := json.MarshalIndent(jsonParsed.Data(), "", "\t")
	os.WriteFile("databases/"+*location+"/database.json", *&jsonData, 0644)
	databases.Store(*location, jsonParsed.Data())
}

// Terminal Websocket
var clients = make(map[*websocket.Conn]bool)

func Terminal(w http.ResponseWriter, r *http.Request) {
	ws, _ := upgrader.Upgrade(w, r, nil)
	defer ws.Close()
	clients[ws] = true

	for {
		var msg string
		err := ws.ReadJSON(&msg)
		if err != nil {
			fmt.Println(err)
			delete(clients, ws)
			break
		}
	}
}

func webparse(ws websocket.Conn) {
	_, message, _ := ws.ReadMessage()
	var set settings
	portgrab(&set)

	fields := strings.Fields(string(*&message))

	if fields[0] == set.Spass {
		if fields[1] == "help" {
			C.help()
		}
	}

}

func mainTerm() {
	for true {
		x := C.term()
		sendtoclients(C.GoString(*&x))
		C.free(unsafe.Pointer(x))
	}
}

func sendtoclients(output string) {
	fmt.Print(*&output)
	for client := range clients {
		client.WriteMessage(websocket.TextMessage, []byte(*&output))
	}
}
