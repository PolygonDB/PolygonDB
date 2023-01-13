package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"

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
	file, _ := ioutil.ReadFile("settings.json")
	var set settings
	json.Unmarshal(file, &set)

	http.HandleFunc("/ws", datahandler)
	log.Println("Server started on -> " + set.Addr + ":" + set.Port)
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

		dbfilename, _ := msg["dbname"].(string) //name of database
		var confdata config = cjson(dbfilename)
		var database map[string]interface{} = djson(dbfilename)

		if msg["password"] == confdata.Key {
			direct := msg["location"].(string)
			action := msg["action"].(string)
			fmt.Print(action, "\n")

			if action == "retrieve" {
				data := retrieve(direct, database)
				ws.WriteJSON(data)
			} else if action == "record" {
				state2 := record(direct, database, msg["value"].(string), dbfilename)
				msg["success"] = state2
				ws.WriteJSON(msg)
			}
		}
	}
}

// Config and Database Parser
func cjson(location string) config {
	file, err := ioutil.ReadFile("databases/" + location + "/config.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
	}

	// Unmarshal the JSON data into a variable
	var jsonData config
	err = json.Unmarshal(file, &jsonData)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
	}
	return jsonData
}

func djson(floc string) map[string]interface{} {
	file, err := ioutil.ReadFile("databases/" + floc + "/database.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
	}

	// Unmarshal the JSON data into a variable
	var jsonData map[string]interface{}
	err = json.Unmarshal(file, &jsonData)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
	}

	// Access the data
	return jsonData
}

// Types of Actions
func retrieve(direct string, database map[string]interface{}) interface{} {
	if direct == "" {
		return database

	} else {

		parts := strings.Split(strings.TrimSpace(direct), "->")

		if len(parts) == 1 {
			data := database[direct]
			return data

		} else {

			//map[string]interface{} -> Json Array
			//[]interfance{} -> Json List

			data := database[parts[0]]
			parts = parts[1:]

			for _, value := range parts {

				x := reflect.TypeOf(data)

				if x.Kind() == reflect.Map { //map[string]interface{}
					data = data.(map[string]interface{})[value]
				} else if x.Kind() == reflect.Slice { //[]interfance{} -> Json List
					i, err := strconv.Atoi(value)
					if err != nil {
						fmt.Println(err)
					} else {
						data = data.([]interface{})[i]
					}
				}
			}
			return data
		}
	}
}

func record(direct string, database map[string]interface{}, value string, location string) string {

	database[direct] = value

	jsonData, _ := json.MarshalIndent(data, "", "\t")
	file, _ := os.Create("databases/" + location + "/database.json")
	defer file.Close()
	file.Write(jsonData)
	return "Finish!"
}
