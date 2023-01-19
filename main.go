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
	Key string `json:"key"`
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
	fmt.Print("Server started on -> "+set.Addr+":"+set.Port, "\n")
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
		defer DBNil(&msg)

		var confdata config
		var database map[string]interface{}
		dbfilename := msg["dbname"].(string)
		er := cd(&dbfilename, &confdata, &database)
		if er != nil {
			ws.WriteJSON("{Error: " + er.Error() + ".}")
		}
		defer DBNil(&database)

		if msg["password"] != confdata.Key {
			ws.WriteJSON("{Error: Password Error.}")
			continue
		}

		direct := msg["location"].(string)
		action := msg["action"].(string)

		if action == "retrieve" {
			data := retrieve(&direct, &database)
			ws.WriteJSON(data)
		} else if action == "record" {
			value := []byte(msg["value"].(string))
			state2 := record(&direct, &database, &value, &dbfilename)
			ws.WriteJSON("{Status: " + state2 + "}")
		} else if action == "search" {
			value := []byte(msg["value"].(string))
			data := search(&direct, &database, &value)
			ws.WriteJSON(data)
		} else if action == "append" {
			fmt.Print("appending...")
			value := []byte(msg["value"].(string))
			data := append(&direct, &database, &value, &dbfilename)
			ws.WriteJSON(data)
		}
		defer StrNil(&action)
		defer StrNil(&direct)
	}
}

// Config and Database Getting
func cd(location *string, jsonData *config, database *map[string]interface{}) error {
	file, err := os.ReadFile("databases/" + *location + "/config.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return err
	}

	// Unmarshal the JSON data into a variable
	err = json.Unmarshal(file, &jsonData)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return err
	}

	file, err = os.ReadFile("databases/" + *location + "/database.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return err
	}

	// Unmarshal the JSON data into a variable
	err = json.Unmarshal(file, &database)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return err
	}

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
	jsonParsed.SetP(val, *direct)
	go Nilify(&val)

	jsonData, _ := json.MarshalIndent(jsonParsed.Data(), "", "\t")
	os.WriteFile("databases/"+*location+"/database.json", jsonData, 0644)

	return "Success"
}

func search(direct *string, database *map[string]interface{}, value *[]byte) interface{} {
	parts := strings.Split(string(*value), ":")
	var output interface{}
	go ByteNil(value)

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

func append(direct *string, database *map[string]interface{}, value *[]byte, location *string) string {

	jsonParsed := parsedata(*database)

	jsonValueParsed, _ := gabs.ParseJSON(*value)
	jsonParsed.ArrayAppendP(jsonValueParsed.Data(), *direct)

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
}

func DBNil(v *map[string]interface{}) {
	*v = nil
}

func ByteNil(v *[]byte) {
	*v = nil
}

func StrNil(v *string) {
	*v = ""
}
