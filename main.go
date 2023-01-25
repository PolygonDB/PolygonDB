package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/Jeffail/gabs/v2"

	"github.com/gorilla/websocket"
)

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

	//go clean()
	http.HandleFunc("/ws", datahandler)
	fmt.Print("Server started on -> "+set.Addr+":"+set.Port, "\n")
	http.ListenAndServe(set.Addr+":"+set.Port, nil)
}

func portgrab(set *settings) {
	file, _ := os.ReadFile("settings.json")
	json.Unmarshal(file, &set)
	file = nil
}

// The GC doesn't work effectively for Websockets for a manual GC is used to help control memory
//func clean() {
//	for {
//		time.Sleep(2 * time.Second)
//		runtime.GC()
//	}
//}

// data handler
var msg map[string]interface{}
var upgrader = websocket.Upgrader{
	EnableCompression: true,
}

func datahandler(w http.ResponseWriter, r *http.Request) {

	ws, _ := upgrader.Upgrade(w, r, nil)
	defer ws.Close()

	for {
		//Reads input
		ws.ReadJSON(&msg)
		if msg == nil {
			break
		}

		process := func(msg *map[string]interface{}) {
			var confdata config
			var database map[string]interface{}

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
				Nullify(&value)
			}

			//When the request is done, it sets everything to either nil or nothing. Easier for GC.
			Nullify(&database)
			Nullify(&confdata)
		}

		process(&msg)
		Nullify(&msg)
		//Cleans up
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
		go fmt.Println("Error unmarshalling Config JSON:", err)
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
		go fmt.Println("Error unmarshalling Database JSON:", err)
		return *err
	}
	Nullify(&file)

	return nil
}

// Types of Actions
func retrieve(direct *string, database *map[string]interface{}) interface{} {

	jsonParsed := parsedata(&*database)

	if *direct == "" {
		return jsonParsed.String()
	} else {
		return jsonParsed.Path(*direct).String()
	}

}

func record(direct *string, database *map[string]interface{}, value *[]byte, location *string) string {

	val, err := UnmarshalJSONValue(&*value)
	if err != nil {
		return "Failure. Value cannot be unmarshal to json."
	}
	go Nullify(&value)

	jsonParsed := parsedata(&database)
	_, er := jsonParsed.SetP(&val, *direct)
	if er != nil {
		return "Failure. Value cannot be placed into database."
	}
	go Nullify(&val)

	jsonData, _ := json.MarshalIndent(jsonParsed.Data(), "", "\t")
	os.WriteFile("databases/"+*location+"/database.json", jsonData, 0644)

	Nullify(&jsonData)
	Nullify(&location)
	Nullify(&jsonParsed)

	return "Success"
}

func search(direct *string, database *map[string]interface{}, value *[]byte) interface{} {
	parts := strings.Split(string(*value), ":")

	var output interface{}
	go Nullify(&value)

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

	val, err := UnmarshalJSONValue(&*value)
	if err != nil {
		return "Failure. Value cannot be unmarshal to json."
	}
	go Nullify(&value)

	er := jsonParsed.ArrayAppendP(val, *direct)
	if er != nil {
		return "Failure!"
	}

	jsonData, _ := json.MarshalIndent(jsonParsed.Data(), "", "\t")

	os.WriteFile("databases/"+*location+"/database.json", jsonData, 0644)

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
	jsonData, _ := json.Marshal(database)
	go Nullify(&database)
	jsonParsed, _ := gabs.ParseJSON([]byte(string(jsonData)))
	return *jsonParsed
}

// Nullify basically helps with the memory management when it comes to websockets
var count int

func Nullify(ptr interface{}) {
	count++
	val := reflect.ValueOf(ptr)
	if val.Kind() == reflect.Ptr {
		val.Elem().Set(reflect.Zero(val.Elem().Type()))
	}
	//count is used to make sure that GC isn't constantly spammed
	if count > 2 {
		runtime.GC()
		count = 0
	}
}
