package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Jeffail/gabs/v2"

	"github.com/gorilla/websocket"

	"github.com/bytedance/sonic"
)

var (
	//uses the json-iterator library since it's faster
	//json = jsoniter.ConfigCompatibleWithStandardLibrary

	//sync/atomic helps with re-using databases so it doesn't constantly re-open a database file
	databases = &atomicDatabase{}

	//Local_only. Can the server be reached outside or only from a certain server?

	upgrader = websocket.Upgrader{
		EnableCompression: true,
		ReadBufferSize:    0,
		WriteBufferSize:   0,
	}

	queue = make(chan wsMessage, 100)

	msg   input
	mutex = &sync.Mutex{}
)

type config struct {
	Key string `json:"key"`
}

type settings struct {
	Addr string `json:"addr"`
	Port string `json:"port"`
	//Lbool bool   `json:"local_only"`
}

// main
func main() {
	var set settings
	portgrab(&set)

	http.HandleFunc("/ws", datahandler)
	fmt.Print("Server started on -> "+set.Addr+":"+set.Port, "\n")
	go mainTerm()
	go processQueue(queue)
	http.ListenAndServe(set.Addr+":"+set.Port, nil)
}

func portgrab(set *settings) {
	file, _ := os.ReadFile("settings.json")
	sonic.Unmarshal(file, &set)
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

type wsMessage struct {
	ws  *websocket.Conn
	msg input
}

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

		if err := sonic.Unmarshal(*&message, &msg); err != nil {
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
			err, output := record(&msg.Loc, &database, &value, &msg.Dbname)
			if err != nil {
				ws.WriteJSON("{Error: " + err.Error() + "}")
			} else {
				ws.WriteJSON("{Status: " + output + "}")
			}

		} else if msg.Act == "search" {
			output := search(&msg.Loc, &database, &value)
			ws.WriteJSON(&output)
		} else if msg.Act == "append" {
			output := append(&msg.Loc, &database, &value, &msg.Dbname)
			ws.WriteJSON("{Status: " + output + "}")
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

		er := datacheck(*&location, *&database)
		if er != nil {
			return er
		} else {
			return nil
		}
	} else {
		return err
	}
}

func datacheck(location *string, database *gabs.Container) error {
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

	content, _ := os.ReadFile("databases/" + *location + "/config.json")

	// Unmarshal the JSON data for config
	*err = sonic.Unmarshal(*&content, &jsonData)

	//*err = json.NewDecoder(file).Decode(&jsonData)
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

func record(direct *string, jsonParsed *gabs.Container, value *[]byte, location *string) (error, string) {

	val, err := UnmarshalJSONValue(&*value)
	if err != nil {
		return err, ""
	}

	_, err = jsonParsed.SetP(&val, *direct)
	if err != nil {
		return err, ""
	}

	syncupdate(jsonParsed, *&location)

	return nil, "Success"
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
			return output
		}
	}

	return "Cannot find value."
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
		err = sonic.Unmarshal(*data, &v)
	case '[':
		if (*data)[len(*data)-1] != ']' {
			return nil, fmt.Errorf("json array is not properly formatted")
		}
		err = sonic.Unmarshal(*data, &v)
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
	jsonData, _ := sonic.ConfigDefault.MarshalIndent(jsonParsed.Data(), "", "    ")
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
	}
	data, _ := sonic.ConfigDefault.MarshalIndent(defaultset, "", "    ")
	os.WriteFile("settings.json", *&data, 0644)
	fmt.Print("Settings.json has been setup. \n")
}

//Embeddable Section
//If the code is being used to embed into another Go Lang project then these functions are designed to that.
//This re-uses the code shown above but re-purposes certain functions for an embed. project

// Starts Polygon Server
func startpolygon(target string) error {
	http.HandleFunc("/ws", datahandler)
	go processQueue(queue)
	er := http.ListenAndServe(target, nil)
	if er != nil {
		return er
	} else {
		fmt.Print("Server started on -> "+target, "\n")
		return nil
	}
}

// Creates a database for you
func create_polygon(name, password *string) error {
	if _, err := os.Stat("databases/" + *name); !os.IsNotExist(err) {
		datacreate(*name, *password)
		return nil
	} else {
		return err
	}
}

// dbname = Name of the Database you are trying to retrieve
// location = Location inside the Database
func polygon_retrieve(dbname string, location string) (error, any) {
	var database gabs.Container
	er := datacheck(&dbname, &database)
	if er != nil {
		return er, nil
	}
	output := retrieve(&location, &database)
	return nil, output
}

func polygon_record(dbname string, location string, value []byte) (error, any) {
	var database gabs.Container
	er := datacheck(&dbname, &database)
	if er != nil {
		return er, nil
	}
	er, output := record(&location, &database, &value, &dbname)
	if er != nil {
		return er, nil
	} else {
		return nil, output
	}
}

func polygon_search(dbname string, location string, value []byte) (error, any) {
	var database gabs.Container
	er := datacheck(&dbname, &database)
	if er != nil {
		return er, nil
	}
	output := search(&location, &database, &value)
	return nil, output
}

func polygon_append(dbname string, location string, value []byte) (error, any) {
	var database gabs.Container
	er := datacheck(&dbname, &database)
	if er != nil {
		return er, nil
	}
	output := append(&location, &database, &value, &location)
	return nil, output
}

type polygon struct {
	data gabs.Container
	name string
}

// If a user wants a "polygon" database and from there modify that, then they can use the following commands:
func get_polygon(dbname string) (error, polygon) {
	var database polygon
	er := datacheck(&dbname, &database.data)
	if er != nil {
		return er, database
	}
	database.name = dbname
	return nil, database
}

func (g polygon) retrieve(location *string) any {
	output := retrieve(location, &g.data)
	return output
}

func (g polygon) record(location *string, value *[]byte) any {
	_, output := record(location, &g.data, value, &g.name)
	return output
}

func (g polygon) search(location *string, value *[]byte) any {
	output := search(location, &g.data, value)
	return output
}

func (g polygon) append(location *string, value *[]byte) any {
	output := append(location, &g.data, value, &g.name)
	return output
}
