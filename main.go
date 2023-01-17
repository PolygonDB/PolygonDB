package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Jeffail/gabs/v2"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type config struct {
	Path string `json:"path"`
	Key  string `json:"key"`
}

type settings struct {
	Addr string `json:"addr"`
	Port string `json:"port"`
}

//type input struct {
//	PassW  string `json:"password"`
//	Dbname string `json:"dbname"`
//	Loc    string `json:"location"`
//	Act    string `json:"action"`
//	Value  string `json:"value"`
//}

func main() {

	file, _ := os.ReadFile("settings.json")
	var set settings
	json.Unmarshal(file, &set)

	http.HandleFunc("/ws", datahandler)
	fmt.Print("Server started on -> " + set.Addr + ":" + set.Port)
	http.ListenAndServe(set.Addr+":"+set.Port, nil)
}

// data handler
func datahandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	for {
		var msg map[string]interface{}
		ws.ReadJSON(&msg)
		if msg == nil {
			break
		}

		var confdata config
		var database map[string]interface{}
		dbfilename := msg["dbname"].(string)
		cd(&dbfilename, &confdata, &database)

		if msg["password"] == confdata.Key {
			direct := msg["location"].(string)
			action := msg["action"].(string)

			if action == "retrieve" {
				data := retrieve(&direct, &database)
				ws.WriteJSON(data)
			} else if action == "record" {
				state2 := record(&direct, &database, []byte(msg["value"].(string)), &dbfilename)
				ws.WriteJSON("{Status: " + state2 + "}")
			} else if action == "search" {
				data := search(&direct, &database, []byte(msg["value"].(string)))
				ws.WriteJSON(data)
			}
		}
	}
}

func cd(location *string, jsonData *config, database *map[string]interface{}) {
	file, err := os.ReadFile("databases/" + *location + "/config.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
	}

	// Unmarshal the JSON data into a variable
	err = json.Unmarshal(file, &jsonData)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
	}

	file, err = os.ReadFile("databases/" + *location + "/database.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
	}

	// Unmarshal the JSON data into a variable
	err = json.Unmarshal(file, &database)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
	}
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

func record(direct *string, database *map[string]interface{}, value []byte, location *string) string {

	val, err := UnmarshalJSONValue(value)
	if err != nil {
		return "Failure"
	}

	jsonParsed := parsedata(*database)
	jsonParsed.SetP(val, *direct)
	go Nilify(&val)

	jsonData, _ := json.MarshalIndent(jsonParsed.Data(), "", "\t")
	os.WriteFile("databases/"+*location+"/database.json", jsonData, 0644)

	return "Success"
}

func search(direct *string, database *map[string]interface{}, value []byte) interface{} {
	parts := strings.Split(string(value), ":")
	var output interface{}

	jsonParsed := parsedata(*database)

	it := jsonParsed.Path("rows").Children()
	for _, user := range it {
		name := user.Path(parts[0]).String()

		if name == parts[1] {
			output = user.Data()
			break
		}
	}

	return output
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
}
