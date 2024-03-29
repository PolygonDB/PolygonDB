package main

import (
	"bufio"
	"context"
	"crypto/rc4"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JewishLewish/PolygonDB/GoPackage/gabs.Revisioned"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"github.com/bytedance/sonic"
)

var (
	databases = &atomicDatabase{
		data: make(map[string][]byte),
	}

	queue = make(chan wsMessage, 100)

	mutex  = &sync.Mutex{}
	logb   bool
	lock   string
	locked = false
	ctx    = context.Background()
)

// Config for databases only holds key
type config struct {
	Key string `json:"key"`
	Enc bool   `json:"encrypted"`
}

// Settings.json parsing
type settings struct {
	Addr string `json:"addr"`
	Port string `json:"port"`
	Logb bool   `json:"log"`
}

// main
// When using a Go Package. This will be ignored. This code is designed for the standalone executable
func main() {

	var set settings = portgrab()

	http.HandleFunc("/ws", datahandler)
	fmt.Print("Server started on -> "+set.Addr+":"+set.Port, "\n")

	go mainterm()
	go processQueue()
	logb = set.Logb

	http.ListenAndServe(set.Addr+":"+set.Port, nil)
}

// Parses the data
// Grabs the informatin from settings.json
func portgrab() settings {
	if _, err := os.Stat("settings.json"); os.IsNotExist(err) {
		setup()
	}

	if _, err := os.Stat("databases"); os.IsNotExist(err) {
		err = os.Mkdir("databases", 0755)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Folder 'databases' created successfully.")
	}

	var set settings
	sonic.Unmarshal(getFilecontent("settings.json"), &set)
	return set
}

func getFilecontent(filename string) []byte {
	file, _ := os.ReadFile("settings.json")
	return file
}

// Uses Atomic Sync for Low Level Sync Pooling and High Memory Efficiency
// Instead of Constantly Re-opening the database json file, this would save the database once and re-use it
type atomicDatabase struct {
	data map[string][]byte
	mu   sync.RWMutex
}

func (ad *atomicDatabase) Load(location string) ([]byte, bool) {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	value, ok := ad.data[location]
	if !ok {
		return nil, false
	}

	return value, true
}

func (ad *atomicDatabase) Store(location string, value []byte) {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	ad.data[location] = value
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

func log(r *http.Request, msg input) {
	output, _ := sonic.ConfigDefault.MarshalIndent(&msg, "", "    ")

	f, err := os.OpenFile("History.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err = f.WriteString(fmt.Sprintf("%s - %s\n", time.Now().String(), "\n\tAddress: "+r.RemoteAddr+"\n\tContent:"+string(output)+"\n")); err != nil {
		panic(err)
	}
}

func datahandler(w http.ResponseWriter, r *http.Request) {

	ws, _ := websocket.Accept(w, r, nil)
	defer ws.Close(websocket.StatusNormalClosure, "")

	if address(&r.RemoteAddr) {
		for {
			if !takein(ws, r) {
				ws.Close(websocket.StatusInternalError, "")
				break
			}
		}
	} else {
		ws.Close(websocket.StatusNormalClosure, "")
	}

}

func address(r *string) bool {
	host, _, _ := net.SplitHostPort(*r)
	if host == "[::1]" {
		return true
	} else {
		return false
	}
}

//Take in takes in the Websocket Message
/*\
From there it does checking to see if it's a valid message or not. If it's not then the for loop for that specific request breaks off.
*/
var msg input

func takein(ws *websocket.Conn, r *http.Request) bool {

	//Reads input
	_, reader, err := ws.Reader(ctx)
	if err != nil {
		return false
	}

	message, _ := io.ReadAll(reader)

	mutex.Lock()
	if err = sonic.Unmarshal(message, &msg); err != nil {
		return false
	}

	//add message to the queue
	queue <- wsMessage{ws: ws, msg: msg}
	mutex.Unlock()
	if logb {
		log(r, msg)
	}

	return true
}

// Processes the Queue. One at a time.
// Both Websocket Handler and Processes Queue work semi-independently
// a Mutex.Lock() is made so it can prevent any possible global variable manipulation and ensures safety
func processQueue() {
	var output wsMessage
	for {
		output = <-queue
		mutex.Lock()
		process(&output.msg, output.ws)
		mutex.Unlock()
	}
}

// Processes the request
// Once request is done, it cleans up out-of-scope variables
var confdata config
var database gabs.Container

func process(msg *input, ws *websocket.Conn) {

	err := cd(&msg.Dbname, &confdata, &database)
	if err != nil {
		wsjson.Write(ctx, ws, `{"Error": "`+err.Error()+`".}`)
		return
	}
	if msg.Pass != confdata.Key {
		wsjson.Write(ctx, ws, `{"Error": "Incorrect Password"}`)
		return
	}

	if msg.Act == "retrieve" {
		wsjson.Write(ctx, ws, retrieve(&msg.Loc, &database))
	} else if msg.Act == "remove" {
		wsjson.Write(ctx, ws, `{"Status": "`+record(&msg.Loc, &database, []byte(""), &msg.Dbname)+`"}`)
	} else {
		if msg.Act == "record" {
			wsjson.Write(ctx, ws, `{"Status": "`+record(&msg.Loc, &database, []byte(fmt.Sprintf("%v", msg.Val)), &msg.Dbname)+`"}`)
		} else if msg.Act == "search" {
			wsjson.Write(ctx, ws, search(&msg.Loc, &database, []byte(fmt.Sprintf("%v", msg.Val))))
		} else if msg.Act == "index" {
			wsjson.Write(ctx, ws, indexsearch(&msg.Loc, &database, []byte(fmt.Sprintf("%v", msg.Val))))
		} else if msg.Act == "append" {
			wsjson.Write(ctx, ws, `{"Status": "`+append_p(&msg.Loc, &database, []byte(fmt.Sprintf("%v", msg.Val)), &msg.Dbname)+`"}`)
		}
	}

	//Cleans up any out-of-scope variables
	defer runtime.GC()
}

// Config and Database Getting
// Uses Concurrency to speed up this process and more precised error handling
func cd(location *string, jsonData *config, database *gabs.Container) error {
	if _, err := os.Stat("databases/" + *location); !os.IsNotExist(err) {

		err = conf(location, jsonData)
		if err != nil {
			return err
		}

		if jsonData.Enc { //if encrypted
			decrypt(location)
			err = datacheck(location, database)
			encrypt(location)
		} else {
			err = datacheck(location, database)
		}

		if err != nil {
			return err
		} else {
			return nil
		}
	} else {
		return err
	}
}

func datacheck(location *string, database *gabs.Container) error {
	if value, ok := databases.Load(*location); ok {
		*database, _ = ParseJSON(&value)
		value = nil
	} else {
		var dataerr error
		*database, dataerr = data(location)
		if dataerr != nil {
			return dataerr
		}
	}
	return nil
}

// This gets the database file
func data(location *string) (gabs.Container, error) {

	value, err := ParseJSONFile("databases/" + *location + "/database.json")
	if err != nil {
		return *value, err
	}
	databases.Store(*location, value.Bytes())
	return *value, nil
}

func conf(location *string, jsonData *config) error {

	content, _ := os.ReadFile("databases/" + *location + "/config.json")

	// Unmarshal the JSON data for config
	err := sonic.Unmarshal(content, &jsonData)

	//*err = json.NewDecoder(file).Decode(&jsonData)
	if err != nil {
		return err
	}
	return nil
}

// Types of Actions
func retrieve(direct *string, jsonParsed *gabs.Container) interface{} {
	if *direct == "" {
		return jsonParsed.Data()
	} else {
		return jsonParsed.Path(*direct).Data()
	}
}

func record(direct *string, jsonParsed *gabs.Container, value []byte, location *string) string {
	if string(value) == "" {
		jsonParsed.DeleteP(*direct)
	} else {
		val, err := unmarshalJSONValue(&value)
		if err != nil {
			return err.Error()
		}

		_, err = jsonParsed.SetP(&val, *direct)

		if err != nil {
			return err.Error()
		}
	}

	syncupdate(jsonParsed, location)

	return "Success"
}

func search(direct *string, jsonParsed *gabs.Container, value []byte) interface{} {
	// Parse the search key and target value
	parts := strings.Split(string(value), ":")
	targ := []byte(parts[1])

	targetValue, _ := unmarshalJSONValue(&targ)

	children := jsonParsed.Path(*direct).Children()
	if int(math.Log2(float64(len(children)))) < 5 {
		return index_s(children, parts[0], targetValue)
	} else {
		return binary_s(children, parts[0], targetValue)
	}
}

func index_s(children []*gabs.Container, searchKey string, targetValue interface{}) interface{} {
	for i, child := range children {
		if child.Path(searchKey).Data() == targetValue {
			return map[string]interface{}{"Index": i, "Value": children[i].Data()}
		}
	}
	return "Cannot find value"
}

func binary_s(children []*gabs.Container, searchKey string, targetValue interface{}) interface{} {
	// Sort the JSON data by the search key
	sort.Slice(children, func(i, j int) bool {
		return fmt.Sprint(children[i].Path(searchKey).Data()) < fmt.Sprint(children[j].Path(searchKey).Data())
	})

	// Perform binary search
	low := 0
	high := len(children) - 1
	for low <= high {
		mid := (low + high) / 2
		midValue := children[mid].Path(searchKey).Data()
		if fmt.Sprint(midValue) == fmt.Sprint(targetValue) {
			return map[string]interface{}{"Index": mid, "Value": children[mid].Data()}
		} else if fmt.Sprint(midValue) < fmt.Sprint(targetValue) {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	return "Cannot find value."
}

func indexsearch(direct *string, jsonParsed *gabs.Container, value []byte) interface{} {
	// Parse the search key and target value
	parts := strings.Split(string(value), ":")
	targ := []byte(parts[1])
	targetValue, _ := unmarshalJSONValue(&targ)

	children := jsonParsed.Path(*direct).Children()
	if int(math.Log2(float64(len(children)))) < 5 {
		return index(children, parts[0], targetValue)
	} else {
		return binary(children, parts[0], targetValue)
	}
}

func index(children []*gabs.Container, searchKey string, targetValue interface{}) interface{} {
	result := make([]interface{}, 0)

	for i, child := range children {
		if child.Path(searchKey).Data() == targetValue {
			result = append(result, map[string]interface{}{"Index": i, "Value": children[i].Data()})
		}
	}
	if len(result) > 0 {
		return result
	} else {
		return "Cannot find value."
	}
}

func binary(children []*gabs.Container, searchKey string, targetValue interface{}) []map[string]interface{} {
	// Make a copy of the original list with the indexes included
	var originalList []map[string]interface{}
	for i, child := range children {
		originalList = append(originalList, map[string]interface{}{"Index": i, "Value": child.Data()})
	}

	// Sort the original list by the search key
	sort.Slice(originalList, func(i, j int) bool {
		return fmt.Sprint(originalList[i]["Value"].(map[string]interface{})[searchKey]) < fmt.Sprint(originalList[j]["Value"].(map[string]interface{})[searchKey])
	})

	// Perform binary search on the sorted list
	low := 0
	high := len(originalList) - 1
	var results []map[string]interface{}
	for low <= high {
		mid := (low + high) / 2
		midValue := originalList[mid]["Value"].(map[string]interface{})[searchKey]
		if fmt.Sprint(midValue) == fmt.Sprint(targetValue) {
			// Collect the matching items
			results = append(results, originalList[mid])
			// Check for other matching items to the left
			for i := mid - 1; i >= low; i-- {
				if fmt.Sprint(originalList[i]["Value"].(map[string]interface{})[searchKey]) == fmt.Sprint(targetValue) {
					results = append(results, originalList[i])
				} else {
					break
				}
			}
			// Check for other matching items to the right
			for i := mid + 1; i <= high; i++ {
				if fmt.Sprint(originalList[i]["Value"].(map[string]interface{})[searchKey]) == fmt.Sprint(targetValue) {
					results = append(results, originalList[i])
				} else {
					break
				}
			}
			sort.Slice(results, func(i, j int) bool {
				return results[i]["Index"].(int) < results[j]["Index"].(int)
			})
			return results
		} else if fmt.Sprint(midValue) < fmt.Sprint(targetValue) {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	return results
}

func append_p(direct *string, jsonParsed *gabs.Container, value []byte, location *string) string {

	val, err := unmarshalJSONValue(&value)
	if err != nil {
		return "Failure. Value cannot be unmarshal to json."
	}

	err = jsonParsed.ArrayAppendP(&val, *direct)
	if err != nil {
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
	if checkenc(location) { //if true...
		decrypt(location)
		WriteFile("databases/"+*location+"/database.json", &jsonData, 0644)
		encrypt(location)
	} else {
		WriteFile("databases/"+*location+"/database.json", &jsonData, 0644)
	}

	databases.Store(*location, jsonParsed.Bytes())
}

// Terminal
func mainterm() {
	scanner := bufio.NewScanner(os.Stdin)
	locked := false
	for {
		scanner.Scan()
		parts := strings.Fields(scanner.Text())
		if len(parts) == 0 {
			continue
		}

		if locked {
			if parts[0] == "unlock" {
				if len(parts) == 1 {
					continue
				} else {
					if parts[1] == lock {
						lock = ""
						locked = false
					}
				}
			} else {
				clearScreen()
			}
		} else {
			switch strings.ToLower(parts[0]) {
			case "help":
				help()
			case "create_database":
				datacreate(&parts[1], &parts[2])
			case "setup":
				setup()
			case "resync":
				resync(&parts[1])
			case "encrypt":
				encrypt(&parts[1])
			case "decrypt":
				decrypt(&parts[1])
			case "change_password":
				chpassword(&parts[1], &parts[2])
			case "lock":
				if len(parts) == 1 {
					continue
				} else {
					lock = parts[1]
					locked = true
					clearScreen()
				}
			}

		}
		parts = nil
	}
}

func help() {
	fmt.Print("\n====Polygon Terminal====\n")
	fmt.Print("help\t\t\t\t\t\tThis displays all the possible executable lines for Polygon\n")
	fmt.Print("create_database (name) (password)\t\tThis will create a database for you with name and password\n")
	fmt.Print("setup\t\t\t\t\t\tCreates settings.json for you\n")
	fmt.Print("resync (name)\t\t\t\t\tRe-syncs a database. For Manual Editing of a database\n")
	fmt.Print("encrypt (name)\t\t\t\t\tEncrypts a database\n")
	fmt.Print("decrypt (name)\t\t\t\t\tDecrypts a database\n")
	fmt.Print("chpassword (name)\t\t\t\tChange password to a database\n")
	fmt.Print("lock (passcode)\t\t\t\t\tLocks the Terminal and Clears Screen\n")
	fmt.Print("unlock (passcode)\t\t\t\tUnlocks Terminal\n")
	fmt.Print("========================\n\n")
}

func datacreate(name, pass *string) {
	path := "databases/" + *name
	os.Mkdir(path, 0777)

	cinput := []byte(fmt.Sprintf(
		`{
	"key": "%s",
	"encrypted": false
}`, *pass))
	WriteFile(path+"/config.json", &cinput, 0644)

	dinput := []byte("{\n\t\"Example\": \"Hello world\"\n}")
	WriteFile(path+"/database.json", &dinput, 0644)
	encrypt(name)

	fmt.Print("File has been created")
}

func chpassword(name, pass *string) {
	content, er := os.ReadFile("databases/" + *name + "/config.json")
	if er != nil {
		fmt.Print(er)
		return
	}
	var conf config
	sonic.Unmarshal(content, &conf)

	if conf.Enc {
		decrypt(name)
	}
	conf.Key = *pass

	content, _ = sonic.ConfigFastest.MarshalIndent(&conf, "", "    ")
	WriteFile("databases/"+*name+"/config.json", &content, 0644)
	encrypt(name)

	fmt.Print("Password successfully changed!")
}

func setup() string {
	defaultset := settings{
		Addr: "0.0.0.0",
		Port: "25565",
		Logb: false,
	}
	data, _ := sonic.ConfigDefault.MarshalIndent(&defaultset, "", "    ")
	WriteFile("settings.json", &data, 0644)
	return "Settings.json has been setup"
}

func resync(name *string) string {
	_, st := databases.Load(*name)
	if !st {
		return "There appears to be no databases previous synced"
	} else {
		decrypt(name)
		defer encrypt(name)
		value, err := ParseJSONFile("databases/" + *name + "/database.json")
		if err != nil {
			return err.Error()
		}
		databases.Store(*name, value.Bytes())
		value = nil
		return "Resync has been successful!"
	}
}

func encrypt(target *string) string {
	var jsonData config
	content, _ := os.ReadFile("databases/" + *target + "/config.json")

	// Unmarshal the JSON data for config
	err := sonic.Unmarshal(content, &jsonData)
	if err != nil {
		return err.Error()
	}

	//if not true...
	if !jsonData.Enc {
		var database gabs.Container
		err = datacheck(target, &database)
		if err != nil {
			return err.Error()
		}

		newtext := deep_encrypt(database.Bytes(), []byte(jsonData.Key))

		jsonData.Enc = true

		output, _ := sonic.ConfigDefault.MarshalIndent(&jsonData, "", "    ")
		WriteFile("databases/"+*target+"/config.json", &output, 0644)
		WriteFile("databases/"+*target+"/database.json", &newtext, 0644)

		return "Encryption successful for " + *target
	} else {
		return "The following data is already encrypted. Don't encrypt again."
	}
}

func decrypt(target *string) string {
	var jsonData config
	content, _ := os.ReadFile("databases/" + *target + "/config.json")

	// Unmarshal the JSON data for config
	err := sonic.Unmarshal(content, &jsonData)
	if err != nil {
		return err.Error()
	}

	//if true...
	if jsonData.Enc {
		database, err := os.ReadFile("databases/" + *target + "/database.json")
		if err != nil {
			return err.Error()
		}

		newtext := deep_decrypt(&database, []byte(jsonData.Key))
		indent(&newtext)
		jsonData.Enc = false

		output, _ := sonic.ConfigDefault.MarshalIndent(&jsonData, "", "    ")
		WriteFile("databases/"+*target+"/config.json", &output, 0644)
		WriteFile("databases/"+*target+"/database.json", &newtext, 0644)

		return "Decryption successful for " + *target
	} else {
		return "Following data is already decrypted. Do not decrypt again."
	}
}

func indent(input *[]byte) {
	var output interface{}
	sonic.ConfigDefault.Unmarshal(*input, &output)
	*input, _ = sonic.ConfigDefault.MarshalIndent(&output, "", "    ")
}

// This code takes normal code from previous functions and uses Ownership + Borrowing
// Memory Efficiency
// $5 Subway
func ParseJSON(sample *[]byte) (gabs.Container, error) {
	var gab interface{}
	if err := sonic.Unmarshal(*sample, &gab); err != nil {
		return *gabs.Wrap(gab), err
	}
	return *gabs.Wrap(gab), nil
}

func ParseJSONFile(path string) (*gabs.Container, error) {

	cBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	container, err := ParseJSON(&cBytes)
	if err != nil {
		return nil, err
	}

	return &container, nil
}

// This is from the OS function. It does the same thing but data now takes in a pointer to make it use less memory
func WriteFile(name string, data *[]byte, perm os.FileMode) error {

	f, err := os.OpenFile(name, 1|64|512, perm)
	if err != nil {
		return err
	}
	_, err = f.Write(*data)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}

func deep_encrypt(plaintext, key []byte) []byte {
	cipher, _ := rc4.NewCipher(key)
	ciphertext := make([]byte, len(plaintext))
	cipher.XORKeyStream(ciphertext, plaintext)
	return ciphertext
}

func deep_decrypt(ciphertext *[]byte, key []byte) []byte {
	cipher, _ := rc4.NewCipher(key)
	plaintext := make([]byte, len(*ciphertext))
	cipher.XORKeyStream(plaintext, *ciphertext)
	return plaintext
}

func checkenc(location *string) bool {
	var jsonData config
	content, _ := os.ReadFile("databases/" + *location + "/config.json")

	// Unmarshal the JSON data for config
	err := sonic.Unmarshal(content, &jsonData)
	if err != nil {
		return false
	}
	return jsonData.Enc
}

// Locking System
func clearScreen() {

	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
	cmd.Run()
	//Runs twice because sometimes pterodactyl servers needs a 2nd clear
}
