package Polygon

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

	polytools "github.com/JewishLewish/PolygonDB/utilities/polyFuncs"
)

var (

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

// Config for databases only holds key
type config struct {
	Key string `json:"key"`
}

// Settings.json parsing
type settings struct {
	Addr string `json:"addr"`
	Port string `json:"port"`
}

// main
// When using a Go Package. This will be ignored. This code is designed for the standalone executable
func main() {
	var set settings
	portgrab(&set)

	http.HandleFunc("/ws", datahandler)
	fmt.Print("Server started on -> "+set.Addr+":"+set.Port, "\n")
	go mainterm()
	go processQueue(queue)
	http.ListenAndServe(set.Addr+":"+set.Port, nil)
}

// Parses the data
// Grabs the informatin from settings.json
func portgrab(set *settings) {
	file, _ := os.ReadFile("settings.json")
	sonic.Unmarshal(file, &set)
	file = nil
}

// Uses Atomic Sync for Low Level Sync Pooling and High Memory Efficiency
// Instead of Constantly Re-opening the database json file, this would save the database once and re-use it
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

// Websocket Message. Each wsMessage is placed in queue
type wsMessage struct {
	ws  *websocket.Conn
	msg input
}

// Parses Input that the Websocket would recieve
type input struct {
	Pass   string `json:"password"`
	Dbname string `json:"dbname"`
	Loc    string `json:"location"`
	Act    string `json:"action"`
	Val    string `json:"value"`
}

// datahandler is where the mainsocker action occurs.
func datahandler(w http.ResponseWriter, r *http.Request) {

	ws, _ := upgrader.Upgrade(w, r, nil)
	defer ws.Close()

	for {
		if !takein(ws) {
			break
		}
	}
}

//Take in takes in the Websocket Message
/*\
From there it does checking to see if it's a valid message or not. If it's not then the for loop for that specific request breaks off.
*/
func takein(ws *websocket.Conn) bool {

	//Reads input
	messageType, reader, err := ws.NextReader()
	if err != nil {
		return false
	}

	switch messageType {
	case websocket.TextMessage:
		message, err := io.ReadAll(reader)
		if err != nil {
			return false
		}

		if err := sonic.Unmarshal(message, &msg); err != nil {
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

// Processes the Queue. One at a time.
// Both Websocket Handler and Processes Queue work semi-independently
// a Mutex.Lock() is made so it can prevent any possible global variable manipulation and ensures safety
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
	defer nullify(&confdata)
	defer nullify(&database)

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
				ws.WriteJSON("{\"Error\": \"" + err.Error() + "\"}")
			} else {
				ws.WriteJSON("{\"Status\": \"" + output + "\"}")
			}

		} else if msg.Act == "search" {
			output := search(&msg.Loc, &database, &value)
			ws.WriteJSON(&output)
		} else if msg.Act == "append" {
			output := append_p(&msg.Loc, &database, &value, &msg.Dbname)
			ws.WriteJSON("{\"Status\": \"" + output + "\"}")
		}
		nullify(&value)
	}

	//When the request is done, it sets everything to either nil or nothing. Easier for GC.
	runtime.GC()
}

// Config and Database Getting
// Uses Concurrency to speed up this process and more precised error handling
func cd(location *string, jsonData *config, database *gabs.Container) error {
	if _, err := os.Stat("databases/" + *location); !os.IsNotExist(err) {
		var conferr error

		conf(&conferr, location, jsonData)
		if conferr != nil {
			return conferr
		}

		er := datacheck(location, database)
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
		*database, _ = polytools.ParseJSON(&value)
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

	value, err := polytools.ParseJSONFile("databases/" + *location + "/database.json")
	if err != nil {
		go fmt.Println("Error unmarshalling Database JSON:", err)
	}
	databases.Store(*location, value.Bytes())
	return nil, *value
}

func conf(err *error, location *string, jsonData *config) {

	content, _ := os.ReadFile("databases/" + *location + "/config.json")

	// Unmarshal the JSON data for config
	*err = sonic.Unmarshal(content, &jsonData)

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
	if string(*value) == "" {
		jsonParsed.DeleteP(*direct)
	} else {
		val, err := unmarshalJSONValue(value)
		if err != nil {
			return err, ""
		}

		_, err = jsonParsed.SetP(&val, *direct)

		if err != nil {
			return err, ""
		}
	}

	syncupdate(jsonParsed, location)

	return nil, "Success"
}

func search(direct *string, jsonParsed *gabs.Container, value *[]byte) interface{} {
	parts := strings.Split(string(*value), ":")
	targ := []byte(parts[1])
	target, _ := unmarshalJSONValue(&targ)
	targ = nil

	var output interface{}

	it := jsonParsed.Path(*direct).Children()
	for i, user := range it {
		if fmt.Sprint(user.Path(parts[0]).Data()) == fmt.Sprint(target) {
			output = map[string]interface{}{"Index": i, "Value": user.Data()}
			return output
		}
	}

	return "Cannot find value."
}

func append_p(direct *string, jsonParsed *gabs.Container, value *[]byte, location *string) string {

	val, err := unmarshalJSONValue(value)
	if err != nil {
		return "Failure. Value cannot be unmarshal to json."
	}

	er := jsonParsed.ArrayAppendP(&val, *direct)
	if er != nil {
		return "Failure!"
	}

	syncupdate(jsonParsed, location)

	return "Success"
}

// Unmarhsals the value into an appropriate json input
func unmarshalJSONValue(data *[]byte) (interface{}, error) {
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
func nullify(ptr interface{}) {
	val := reflect.ValueOf(ptr)
	if val.Kind() == reflect.Ptr {
		val.Elem().Set(reflect.Zero(val.Elem().Type()))
	}
}

// Sync Update
// Since we are using atomic/sync for memory efficiency. We need to make sure that when the atomic database is updated, then we can update the sync database
func syncupdate(jsonParsed *gabs.Container, location *string) {
	jsonData, _ := sonic.ConfigDefault.MarshalIndent(jsonParsed.Data(), "", "    ")
	polytools.WriteFile("databases/"+*location+"/database.json", &jsonData, 0644)
	databases.Store(*location, jsonParsed.Bytes())
}

//Embeddable Section
//If the code is being used to embed into another Go Lang project then these functions are designed to that.
//This re-uses the code shown above but re-purposes certain functions for an embed. project

// Starts Polygon Server
func Start(target string) error {
	http.HandleFunc("/ws", datahandler)
	go processQueue(queue)
	fmt.Print("Server starting on => " + target)
	er := http.ListenAndServe(target, nil)
	if er != nil {
		return er
	} else {
		fmt.Print("Server started on -> "+target, "\n")
		return nil
	}
}

// Creates a database for you
func Create(name, password string) error {
	if _, err := os.Stat("databases"); os.IsNotExist(err) {
		os.Mkdir("databases", 0777)
	}

	if _, err := os.Stat("databases/" + name); err != nil {
		datacreate(name, password)
		return nil
	} else {
		return err
	}
}

// dbname = Name of the Database you are trying to retrieve
// location = Location inside the Database
func Retrieve_P(dbname string, location string) (any, error) {
	var database gabs.Container
	er := datacheck(&dbname, &database)
	if er != nil {
		return nil, er
	}
	output := retrieve(&location, &database)
	return output, nil
}

func Record_P(dbname string, location string, value []byte) (any, error) {
	var database gabs.Container
	er := datacheck(&dbname, &database)
	if er != nil {
		return nil, er
	}
	er, output := record(&location, &database, &value, &dbname)
	if er != nil {
		return nil, er
	} else {
		return output, nil
	}
}

func Search_P(dbname string, location string, value []byte) (any, error) {
	var database gabs.Container
	er := datacheck(&dbname, &database)
	if er != nil {
		return nil, er
	}
	output := search(&location, &database, &value)
	return output, nil
}

func Append_P(dbname string, location string, value []byte) (any, error) {
	var database gabs.Container
	er := datacheck(&dbname, &database)
	if er != nil {
		return nil, er
	}
	output := append_p(&location, &database, &value, &location)
	return output, nil
}

type Polygon struct {
	data gabs.Container
	name string
}

// If a user wants a "polygon" database and from there modify that, then they can use the following commands:
func Get(dbname string) (Polygon, error) {
	var database Polygon
	er := datacheck(&dbname, &database.data)
	if er != nil {
		return database, er
	}
	database.name = dbname
	return database, nil
}

func (g Polygon) Retrieve(location string) any {
	output := retrieve(&location, &g.data)
	return output
}

func (g Polygon) Record(location string, value []byte) any {
	_, output := record(&location, &g.data, &value, &g.name)
	return output
}

func (g Polygon) Search(location string, value []byte) map[string]interface{} {
	output := search(&location, &g.data, &value)
	if output == "Cannot find value." {
		return nil
	} else {
		return output.(map[string]interface{})
	}

}

func (g Polygon) Append(location string, value []byte) any {
	output := append_p(&location, &g.data, &value, &g.name)
	return output
}

// Terminal
// This is designed for the standalone executable.
// However, datacreate() is used in the Create Function for Go Package
func mainterm() {
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

		parts = nil
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
	polytools.WriteFile(conpath, &cinput, 0644)

	datapath := path + "/database.json"
	dinput := []byte("{\n\t\"Example\": \"Hello world\"\n}")
	polytools.WriteFile(datapath, &dinput, 0644)

	fmt.Println("File has been created.")
}

func setup() {
	type settings struct {
		Addr string `json:"addr"`
		Port string `json:"port"`
	}
	defaultset := settings{
		Addr: "0.0.0.0",
		Port: "25565",
	}
	data, _ := sonic.ConfigDefault.MarshalIndent(&defaultset, "", "    ")
	polytools.WriteFile("settings.json", &data, 0644)
	fmt.Print("Settings.json has been setup. \n")
}
