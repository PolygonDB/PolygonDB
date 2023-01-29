package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Jeffail/gabs/v2"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"

	"github.com/gorilla/websocket"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type config struct {
	Key string `json:"key"`
}

type settings struct {
	Addr string `json:"addr"`
	Port string `json:"port"`
	Goe  bool   `json:"goextend"`
	Gof  string `json:"gofile"`
}

// main
func main() {
	var set settings
	portgrab(&set)

	if set.Goe {
		getGo(set.Gof)
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

// Uses Atomic Sync for Low Level Sync Pooling and High Memory Efficiency
type atomicDatabase struct {
	value atomic.Value
}

func (ad *atomicDatabase) Load(location string) ([]byte, bool) {
	v := ad.value.Load()
	if v == nil {
		return nil, false
	}
	value := v.([]byte)
	return value, true
}

func (ad *atomicDatabase) Store(location string, value []byte) {
	ad.value.Store(value)
}

var databases = &atomicDatabase{}

var upgrader = websocket.Upgrader{
	EnableCompression: true,
	ReadBufferSize:    0,
	WriteBufferSize:   0,
}

type wsMessage struct {
	ws  *websocket.Conn
	msg input
}

var queue = make(chan wsMessage, 100)

type input struct {
	Pass   string `json:"password"`
	Dbname string `json:"dbname"`
	Loc    string `json:"location"`
	Act    string `json:"action"`
	Val    string `json:"value"`
}

func datahandler(w http.ResponseWriter, r *http.Request) {

	ws, _ := upgrader.Upgrade(w, r, nil)
	defer ws.Close()

	for {
		if !takein(ws) {
			break
		}
	}
}

var msg input

func takein(ws *websocket.Conn) bool {

	//Reads input
	messageType, reader, err := ws.NextReader()
	if err != nil {
		return false
	}

	switch messageType {
	case websocket.TextMessage:
		message, err := io.ReadAll(*&reader)
		if err != nil {
			return false
		}

		if err := json.Unmarshal(*&message, &msg); err != nil {
			return false
		}

		//add message to the queue
		mutex.Lock()
		queue <- wsMessage{ws: ws, msg: msg}
		mutex.Unlock()
	default:
		return false
	}
	return true
}

var mutex = &sync.Mutex{}

func processQueue(queue chan wsMessage) {
	for {
		msg := <-queue
		mutex.Lock()
		process(&msg.msg, msg.ws)
		mutex.Unlock()
	}
}

// Processes the request
// Once request is done, it cleans up out-of-scope variables
func process(msg *input, ws *websocket.Conn) {

	var confdata config
	var database gabs.Container

	er := cd(&msg.Dbname, &confdata, &database)
	if er != nil {
		ws.WriteJSON("{Error: " + er.Error() + ".}")
		return
	}
	if msg.Pass != confdata.Key {
		ws.WriteJSON("{Error: Password Error.}")
		return
	}
	defer Nullify(&confdata)
	defer Nullify(&database)

	//direct := (*msg)["location"].(string)
	//action := (*msg)["action"].(string)

	if msg.Act == "retrieve" {
		output := retrieve(&msg.Loc, &database)
		ws.WriteJSON(&output)
	} else {
		value := []byte(msg.Val)
		if msg.Act == "record" {
			output := record(&msg.Loc, &database, &value, &msg.Dbname)
			ws.WriteJSON("{Status: " + output + "}")
		} else if msg.Act == "search" {
			output := search(&msg.Loc, &database, &value)
			ws.WriteJSON(&output)
		} else if msg.Act == "append" {
			output := append(&msg.Loc, &database, &value, &msg.Dbname)
			ws.WriteJSON("{Status: " + output + "}")
		} else if msg.Act == "custom" {
			output := doGo(&value, &database, &msg.Loc)
			ws.WriteJSON(output.String())
		}
		Nullify(&value)
	}

	//When the request is done, it sets everything to either nil or nothing. Easier for GC.
	runtime.GC()

}

// Config and Database Getting
// Uses Concurrency to speed up this process and more precised error handling
func cd(location *string, jsonData *config, database *gabs.Container) error {
	if _, err := os.Stat("databases/" + *location); !os.IsNotExist(err) {
		var conferr error

		conf(&conferr, &*location, &*jsonData)
		if conferr != nil {
			return conferr
		}

		if value, ok := databases.Load(*location); ok {
			parsed, _ := gabs.ParseJSON(value)
			*database = *parsed
			value = nil
		} else {
			var dataerr error
			dataerr, *database = data(location)
			if dataerr != nil {
				return dataerr
			}
		}

		return nil

	} else {
		return err
	}
}

// This gets the database file
func data(location *string) (error, gabs.Container) {
	value, err := gabs.ParseJSONFile("databases/" + *location + "/database.json")
	if err != nil {
		go fmt.Println("Error unmarshalling Database JSON:", err)
	}
	databases.Store(*location, value.Bytes())
	return err, *value
}

func conf(err *error, location *string, jsonData *config) {
	var file *os.File
	file, *err = os.Open("databases/" + *location + "/config.json")
	if *err != nil {
		go fmt.Println("Error reading file:", err)
	}
	defer file.Close()

	// Unmarshal the JSON data for config
	*err = json.NewDecoder(file).Decode(&jsonData)
	if *err != nil {
		go fmt.Println("Error unmarshalling Config JSON:", err)
	}
}

// Types of Actions
func retrieve(direct *string, jsonParsed *gabs.Container) interface{} {
	if *direct == "" {
		return jsonParsed.Data()
	} else {
		return jsonParsed.Path(*direct).Data()
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

	syncupdate(jsonParsed, *&location)

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

	syncupdate(jsonParsed, *&location)

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

// Nullify basically helps with the memory management when it comes to websockets
func Nullify(ptr interface{}) {
	val := reflect.ValueOf(*&ptr)
	if val.Kind() == reflect.Ptr {
		val.Elem().Set(reflect.Zero(val.Elem().Type()))
	}
}

// Sync Update
func syncupdate(jsonParsed *gabs.Container, location *string) {
	jsonData, _ := json.MarshalIndent(jsonParsed.Data(), "", "    ")
	os.WriteFile("databases/"+*location+"/database.json", *&jsonData, 0644)
	databases.Store(*location, jsonParsed.Bytes())
}

// Terminal Websocket
func mainTerm() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		parts := strings.Fields(scanner.Text())
		if len(parts) == 0 {
			continue
		}

		if parts[0] == "help" {
			help()
		} else if parts[0] == "create_database" {
			datacreate(parts[1], parts[2])
		} else if parts[0] == "setup" {
			setup()
		}

		Nullify(&parts)
	}
}

func help() {
	fmt.Print("\n====Polygon Terminal====\n")
	fmt.Print("help\t\t\t\t\t\tThis displays all the possible executable lines for Polygon\n")
	fmt.Print("create_database (name) (password)\t\tThis will create a database for you with name and password\n")
	fmt.Print("setup\t\t\t\t\t\tCreates settings.json for you\n")
	fmt.Print("========================\n\n")
}

func datacreate(name, pass string) {
	path := "databases/" + name
	os.Mkdir(path, 0777)

	conpath := path + "/config.json"
	cinput := []byte(fmt.Sprintf("{\n\t\"key\": \"%s\"\n}", pass))
	os.WriteFile(conpath, cinput, 0644)

	datapath := path + "/database.json"
	dinput := []byte("{\n\t\"Example\": \"Hello world\"\n}")
	os.WriteFile(datapath, dinput, 0644)

	fmt.Println("File has been created.")
}

func setup() {
	defaultset := settings{
		Addr: "0.0.0.0",
		Port: "25565",
		Goe:  false,
		Gof:  "link",
	}
	data, _ := json.MarshalIndent(defaultset, "", "    ")
	os.WriteFile("settings.json", *&data, 0644)
	fmt.Print("Settings.json has been setup. \n")
}

var i *interp.Interpreter = interp.New(interp.Options{})

func getGo(loc string) {
	resp, err := http.Get(loc)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	i.Use(stdlib.Symbols)

	_, err = i.Eval(`` + string(body) + ``)
	if err != nil {
		println(err)
	}

	fmt.Print("Succesfully parsed the data! \n")
}

func doGo(target *[]byte, database *gabs.Container, direct *string) reflect.Value {
	parbyte := string(*target)
	if strings.Contains(parbyte, "{database}") {
		mut := retrieve(*&direct, *&database).(string)
		*&mut = strings.Replace(mut, "\"", "'", -1)
		*&parbyte = strings.Replace(parbyte, "{database}", "\""+mut+"\"", -1)
	}

	v, err := i.Eval(`` + parbyte + ``)
	if err != nil {
		fmt.Print(err)
	}
	return v
}
