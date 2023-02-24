package main

import (
	"bufio"
	"context"
	"crypto/rc4"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"runtime"
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

	mutex     = &sync.Mutex{}
	whitelist []interface{}
	logb      bool
	lock      string
	ctx       = context.Background()
)

// Config for databases only holds key
type config struct {
	Key string `json:"key"`
	Enc bool   `json:"encrypted"`
}

// Settings.json parsing
type settings struct {
	Addr     string        `json:"addr"`
	Port     string        `json:"port"`
	Logb     bool          `json:"log"`
	Whiteadd []interface{} `json:"whitelist_addresses"`
}

// main
// When using a Go Package. This will be ignored. This code is designed for the standalone executable
func main() {
	var set settings
	portgrab(&set)

	http.HandleFunc("/ws", datahandler)
	fmt.Print("Server started on -> "+set.Addr+":"+set.Port, "\n")

	go mainterm()
	go processQueue()
	logb = set.Logb
	whitelist = set.Whiteadd

	http.ListenAndServe(set.Addr+":"+set.Port, nil)
}

// Parses the data
// Grabs the informatin from settings.json
func portgrab(set *settings) {
	if _, err := os.Stat("settings.json"); os.IsNotExist(err) {
		setup()
	}

	file, _ := os.ReadFile("settings.json")
	sonic.Unmarshal(file, &set)
	file = nil

	if _, err := os.Stat("databases"); os.IsNotExist(err) {
		err = os.Mkdir("databases", 0755)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Folder 'databases' created successfully.")
	}
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
	data := "\n\tAddress: " + r.RemoteAddr + "\n\tContent:" + string(output) + "\n"

	f, err := os.OpenFile("History.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err := f.WriteString(fmt.Sprintf("%s - %s\n", time.Now().String(), data)); err != nil {
		panic(err)
	}
}

// datahandler is where the mainsocker action occurs.
func datahandler(w http.ResponseWriter, r *http.Request) {

	ws, _ := websocket.Accept(w, r, nil)
	defer ws.Close(websocket.StatusNormalClosure, "")

	if address(&r.RemoteAddr) {
		for {
			if !takein(ws, r) {
				break
			}
		}
	} else {
		ws.Close(websocket.StatusNormalClosure, "")
	}

}

func address(r *string) bool {
	if len(whitelist) == 0 {
		return true
	} else {
		host, _, _ := net.SplitHostPort(*r)
		defer nullify(&host)
		if contains(&whitelist, &host) {
			return true
		} else {
			return false
		}
	}
}

func contains(s *[]interface{}, str *string) bool {
	for _, v := range *s {
		if v == *str {
			return true
		}
	}
	return false
}

//Take in takes in the Websocket Message
/*\
From there it does checking to see if it's a valid message or not. If it's not then the for loop for that specific request breaks off.
*/
func takein(ws *websocket.Conn, r *http.Request) bool {

	//Reads input
	_, reader, err := ws.Read(ctx)
	if err != nil {
		return false
	}

	var msg input
	if err = sonic.Unmarshal(reader, &msg); err != nil {
		return false
	}

	//add message to the queue
	mutex.Lock()
	queue <- wsMessage{ws: ws, msg: msg}
	mutex.Unlock()
	if logb {
		log(r, msg)
	}
	defer nullify(&msg)

	return true
}

// Processes the Queue. One at a time.
// Both Websocket Handler and Processes Queue work semi-independently
// a Mutex.Lock() is made so it can prevent any possible global variable manipulation and ensures safety
func processQueue() {
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

	err := cd(&msg.Dbname, &confdata, &database)
	if err != nil {
		wsjson.Write(ctx, ws, "{Error: "+err.Error()+".}")
		return
	}
	if msg.Pass != confdata.Key {
		wsjson.Write(ctx, ws, "{Error: Password Error.}")
		return
	}
	defer nullify(&confdata)
	defer nullify(&database)

	if msg.Act == "retrieve" {
		wsjson.Write(ctx, ws, retrieve(&msg.Loc, &database))
	} else {
		value := []byte(msg.Val)
		if msg.Act == "record" {
			output, err := record(&msg.Loc, &database, &value, &msg.Dbname)
			if err != nil {
				wsjson.Write(ctx, ws, "{\"Error\": \""+err.Error()+"\"}")
			} else {
				wsjson.Write(ctx, ws, "{\"Status\": \""+output+"\"}")
			}

		} else if msg.Act == "search" {
			output := search(&msg.Loc, &database, &value)
			wsjson.Write(ctx, ws, &output)
		} else if msg.Act == "append" {
			output := append_p(&msg.Loc, &database, &value, &msg.Dbname)
			wsjson.Write(ctx, ws, "{\"Status\": \""+output+"\"}")
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
	databases.Store(*location, GabtoBytes(value))
	return *value, nil
}

func GabtoBytes(g *gabs.Container) []byte {
	if bytes, err := sonic.ConfigFastest.Marshal(g.Data()); err == nil {
		return bytes
	}
	return []byte("null")
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

func record(direct *string, jsonParsed *gabs.Container, value *[]byte, location *string) (string, error) {
	if string(*value) == "" {
		jsonParsed.DeleteP(*direct)
	} else {
		val, err := unmarshalJSONValue(value)
		if err != nil {
			return "", err
		}

		_, err = jsonParsed.SetP(&val, *direct)

		if err != nil {
			return "", err
		}
	}

	syncupdate(jsonParsed, location)

	return "Success", nil
}

func search(direct *string, jsonParsed *gabs.Container, value *[]byte) interface{} {
	parts := strings.Split(string(*value), ":")
	targ := []byte(parts[1])
	target, _ := unmarshalJSONValue(&targ)
	targ = nil

	for i, user := range jsonParsed.Path(*direct).Children() {
		if user.Path(parts[0]).Data() == target {
			return map[string]interface{}{"Index": i, "Value": user.Data()}
		}
	}

	return "Cannot find value."
}

func append_p(direct *string, jsonParsed *gabs.Container, value *[]byte, location *string) string {

	val, err := unmarshalJSONValue(value)
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

//Embeddable Section
//If the code is being used to embed into another Go Lang project then these functions are designed to that.
//This re-uses the code shown above but re-purposes certain functions for an embed. project

// Starts Polygon Server
func Start(target string) error {
	http.HandleFunc("/ws", datahandler)
	go processQueue()
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
		datacreate(&name, &password)
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
	output, er := record(&location, &database, &value, &dbname)
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

	fmt.Println("File has been created.")
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
		fmt.Print("Turn off encryption first before changing password as it can break the database!\n")
		return
	}
	conf.Key = *pass

	content, _ = sonic.ConfigFastest.MarshalIndent(&conf, "", "    ")
	WriteFile("databases/"+*name+"/config.json", &content, 0644)

	fmt.Print("Password successfully changed!\n")
}

func setup() {
	defaultset := settings{
		Addr:     "0.0.0.0",
		Port:     "25565",
		Logb:     false,
		Whiteadd: make([]interface{}, 0),
	}
	data, _ := sonic.ConfigFastest.MarshalIndent(&defaultset, "", "    ")
	WriteFile("settings.json", &data, 0644)
	fmt.Print("Settings.json has been setup. \n")
}

func resync(name *string) {
	_, st := databases.Load(*name)
	if !st {
		fmt.Print("There appears to be no databases previous synced...\n")
		return
	} else {
		value, err := ParseJSONFile("databases/" + *name + "/database.json")
		if err != nil {
			fmt.Println("Error unmarshalling Database JSON:", err)
			return
		}
		databases.Store(*name, value.Bytes())
		fmt.Print("Resync has been successful!\n")
		value = nil
	}
}

func encrypt(target *string) {
	var jsonData config
	content, _ := os.ReadFile("databases/" + *target + "/config.json")

	// Unmarshal the JSON data for config
	err := sonic.Unmarshal(content, &jsonData)
	if err != nil {
		fmt.Print(err)
		return
	}

	//if not true...
	if !jsonData.Enc {
		var database gabs.Container
		err = datacheck(target, &database)
		if err != nil {
			fmt.Print(err)
			return
		}

		newtext := deep_encrypt(database.Bytes(), []byte(jsonData.Key))
		fmt.Print(string(newtext), "\n")

		jsonData.Enc = true

		output, _ := sonic.ConfigDefault.MarshalIndent(&jsonData, "", "    ")
		WriteFile("databases/"+*target+"/config.json", &output, 0644)
		WriteFile("databases/"+*target+"/database.json", &newtext, 0644)

		fmt.Print("Encryption successful for " + *target + ".\n")
	} else {
		fmt.Print("The following data is already encrypted. Don't encrypt again.\n")
	}
}

func decrypt(target *string) {
	var jsonData config
	content, _ := os.ReadFile("databases/" + *target + "/config.json")

	// Unmarshal the JSON data for config
	err := sonic.Unmarshal(content, &jsonData)
	if err != nil {
		fmt.Print(err)
		return
	}

	//if true...
	if jsonData.Enc {
		database, err := os.ReadFile("databases/" + *target + "/database.json")
		if err != nil {
			fmt.Print(err, "\n")
			return
		}

		newtext := deep_decrypt(&database, []byte(jsonData.Key))
		indent(&newtext)
		jsonData.Enc = false

		output, _ := sonic.ConfigDefault.MarshalIndent(&jsonData, "", "    ")
		WriteFile("databases/"+*target+"/config.json", &output, 0644)
		WriteFile("databases/"+*target+"/database.json", &newtext, 0644)

		fmt.Print("Decryption successful for " + *target + ".\n")
	} else {
		fmt.Print("Following data is already decrypted. Do not decrypt again.\n")
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
