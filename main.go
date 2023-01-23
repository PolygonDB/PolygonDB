package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jeffail/gabs/v2"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type config struct {
	Key string `json:"key"`
}

type settings struct {
	Addr string `json:"addr"`
	Port string `json:"port"`
}

// main
func main() {

	var set settings
	portgrab(&set)

	go clean()
	http.HandleFunc("/ws", datahandler)
	fmt.Print("Server started on -> "+set.Addr+":"+set.Port, "\n")
	http.ListenAndServe(set.Addr+":"+set.Port, nil)
}

func portgrab(set *settings) {
	file, _ := os.ReadFile("settings.json")
	json.Unmarshal(file, &set)
	file = nil
}

var connectionPool = sync.Pool{
	New: func() interface{} {
		return &websocket.Conn{}
	},
}

func clean() {
	for {
		time.Sleep(5 * time.Second)
		fmt.Print("Cleaning...")
		runtime.GC()
	}

}

// Using Global Variables, since users will request a lot of data, these variables can be used to prevent constant repititon and hurting the Go Lang's GC
var confdata config
var database map[string]interface{}
var msg map[string]interface{}
var direct string
var action string
var value []byte
var dbfilename string
var state string
var data interface{}

// data handler
func datahandler(w http.ResponseWriter, r *http.Request) {
	ws := connectionPool.Get().(*websocket.Conn)
	defer connectionPool.Put(ws)

	ws, _ = upgrader.Upgrade(w, r, nil)
	defer ws.Close()

	for {
		ws.ReadJSON(&msg)
		if msg == nil {
			break
		}
		defer DBNil(&msg)

		dbfilename = msg["dbname"].(string)
		er := cd(&dbfilename, &confdata, &database)
		if er != nil {
			ws.WriteJSON("{Error: " + er.Error() + ".}")
		}
		defer DBNil(&database)

		if msg["password"] != confdata.Key {
			ws.WriteJSON("{Error: Password Error.}")
			continue
		}

		direct = msg["location"].(string)
		action = msg["action"].(string)

		if action == "retrieve" {
			data = retrieve(&direct, &database)
			ws.WriteJSON(data)
		} else if action == "record" {
			value = []byte(msg["value"].(string))
			state = record(&direct, &database, &value, &dbfilename)
			ws.WriteJSON("{Status: " + state + "}")
		} else if action == "search" {
			value = []byte(msg["value"].(string))
			data = search(&direct, &database, &value)
			ws.WriteJSON(data)
		} else if action == "append" {
			value = []byte(msg["value"].(string))
			state = append(&direct, &database, &value, &dbfilename)
			ws.WriteJSON("{Status: " + state + "}")
		}
		action = ""
		direct = ""
		msg = nil
		dbfilename = ""
	}
}

// Config and Database Getting
func cd(location *string, jsonData *config, database *map[string]interface{}) error {
	file := new([]byte)
	err := new(error)
	*file, *err = os.ReadFile("databases/" + *location + "/config.json")
	if *err != nil {
		go fmt.Println("Error reading file:", err)
		return *err
	}

	// Unmarshal the JSON data into a variable
	*err = json.Unmarshal(*file, &jsonData)
	if *err != nil {
		go fmt.Println("Error unmarshalling JSON:", err)
		return *err
	}

	*file, *err = os.ReadFile("databases/" + *location + "/database.json")
	if *err != nil {
		go fmt.Println("Error reading file:", err)
		return *err
	}

	// Unmarshal the JSON data into a variable
	*err = json.Unmarshal(*file, &database)
	if *err != nil {
		go fmt.Println("Error unmarshalling JSON:", err)
		return *err
	}
	*file = nil

	return nil
}

// Types of Actions
func retrieve(direct *string, database *map[string]interface{}) interface{} {

	jsonParsed := parsedata(*database)

	if *direct == "" {
		return jsonParsed.String()
	} else {
		return jsonParsed.Path(*direct).String()
	}

}

func record(direct *string, database *map[string]interface{}, value *[]byte, location *string) string {

	val, err := UnmarshalJSONValue(*value)
	if err != nil {
		return "Failure"
	}
	go ByteNil(value)

	jsonParsed := parsedata(*database)
	_, er := jsonParsed.SetP(&val, *direct)
	if er != nil {
		return "Failure"
	}
	go Nilify(&val)

	jsonData, _ := json.MarshalIndent(jsonParsed.Data(), "", "\t")
	os.WriteFile("databases/"+*location+"/database.json", jsonData, 0644)

	return "Success"
}

func search(direct *string, database *map[string]interface{}, value *[]byte) interface{} {
	parts := strings.Split(string(*value), ":")
	var output interface{}
	output = "Hello world"
	go ByteNil(value)

	jsonParsed := parsedata(*database)

	it := jsonParsed.Path(*direct).Children()
	for i, user := range it {
		name := user.Path(parts[0]).String()

		if name == parts[1] {
			output = map[string]interface{}{"Index": i, "Value": user.Data()}
			break
		}
	}

	return output
}

func append(direct *string, database *map[string]interface{}, value *[]byte, location *string) string {

	jsonParsed := parsedata(*database)

	jsonValueParsed, _ := gabs.ParseJSON(*value)
	er := jsonParsed.ArrayAppendP(jsonValueParsed.Data(), *direct)
	if er != nil {
		return "Failure!"
	}

	jsonData, _ := json.MarshalIndent(jsonParsed.Data(), "", "\t")

	os.WriteFile("databases/"+*location+"/database.json", jsonData, 0644)

	return "Success"
}

func UnmarshalJSONValue(data []byte) (interface{}, error) {
	var v interface{}
	var err error
	if len(data) == 0 {
		return nil, fmt.Errorf("json data is empty")
	}
	switch data[0] {
	case '"':
		if data[len(data)-1] != '"' {
			return nil, fmt.Errorf("json string is not properly formatted")
		}
		v = string(data[1 : len(data)-1])
	case '{':
		if data[len(data)-1] != '}' {
			return nil, fmt.Errorf("json object is not properly formatted")
		}
		err = json.Unmarshal(data, &v)
	case '[':
		if data[len(data)-1] != ']' {
			return nil, fmt.Errorf("json array is not properly formatted")
		}
		err = json.Unmarshal(data, &v)
	default:
		i, e := strconv.Atoi(string(data))
		if e != nil {
			return nil, fmt.Errorf("unable to parse json data")
		}
		v = i
	}
	return v, err
}

func parsedata(database interface{}) gabs.Container {
	jsonData, _ := json.Marshal(database)
	go Nilify(&database)
	jsonParsed, _ := gabs.ParseJSON([]byte(string(jsonData)))
	return *jsonParsed
}

func Nilify(v *interface{}) {
	*v = nil
	runtime.GC()
}

func DBNil(v *map[string]interface{}) {
	*v = nil
	runtime.GC()
}

func ByteNil(v *[]byte) {
	*v = nil
}

func StrNil(v *string) {
	*v = ""
}
