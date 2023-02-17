package terms

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	polySecurity "github.com/JewishLewish/PolygonDB/GoPackage/utilities/polyEncrypt"
	utils "github.com/JewishLewish/PolygonDB/GoPackage/utilities/polyFuncs"
	polygon "github.com/JewishLewish/PolygonDB/GoPackage/utilities/polyStructs"
	polysync "github.com/JewishLewish/PolygonDB/GoPackage/utilities/polySync"
	"github.com/bytedance/sonic"
	"github.com/gorilla/websocket"
)

// Terminal
// This is designed for the standalone executable.
// However, datacreate() is used in the Create Function for Go Package
var (
	locked        = false
	lock   string = ""
)

func Terminal() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		Parse(strings.Fields(scanner.Text()), nil)
	}
}

func Parse(parts []string, ws *websocket.Conn) {
	var output string
	if len(parts) == 0 {
		return
	}

	if locked {
		if parts[0] == "unlock" {
			if len(parts) == 1 {
				return
			} else {
				if parts[1] == lock {
					lock = ""
					locked = false
					output = "Terminal has been unlocked successfully."
				}
			}
		} else {
			clearScreen()
		}
	} else {
		switch strings.ToLower(parts[0]) {
		case "help":
			output = help()
		case "create_database":
			Datacreate(&parts[1], &parts[2])
		case "setup":
			setup()
		case "resync":
			resync(&parts[1])
		case "encrypt":
			polySecurity.Encrypt(&parts[1])
		case "decrypt":
			polySecurity.Decrypt(&parts[1])
		case "change_password":
			chpassword(&parts[1], &parts[2])
		case "lock":
			if len(parts) == 1 {
				return
			} else {
				lock = parts[1]
				locked = true
				clearScreen()
				return
			}
		default:
			return
		}
	}

	if ws != nil {
		ws.WriteJSON(output)
	}
}

func help() string {
	return `
\n====Polygon Terminal====\n
help									This displays all the possible executable lines for Polygon
create_database (name) (password)		This will create a database for you with name and password
setup									Creates settings.json for you\n
resync (name)							Re-syncs a database. For Manual Editing of a database
========================
`
}

func Datacreate(name, pass *string) {
	path := "databases/" + *name
	os.Mkdir(path, 0777)

	cinput := []byte(fmt.Sprintf(
		`{
	"key": "%s",
	"encrypted": false
}`, *pass))
	utils.WriteFile(path+"/config.json", &cinput, 0644)

	dinput := []byte("{\n\t\"Example\": \"Hello world\"\n}")
	utils.WriteFile(path+"/database.json", &dinput, 0644)

	fmt.Println("File has been created.")
}

func chpassword(name, pass *string) {
	content, er := os.ReadFile("databases/" + *name + "/config.json")
	if er != nil {
		fmt.Print(er)
		return
	}
	var conf polygon.Config
	sonic.Unmarshal(content, &conf)

	if conf.Enc {
		fmt.Print("Turn off encryption first before changing password as it can break the database!\n")
		return
	}
	conf.Key = *pass

	content, _ = sonic.ConfigFastest.MarshalIndent(&conf, "", "    ")
	utils.WriteFile("databases/"+*name+"/config.json", &content, 0644)

	fmt.Print("Password successfully changed!\n")
}

func setup() {
	var w []interface{}
	defaultset := polygon.Settings{
		Addr:     "0.0.0.0",
		Port:     "25565",
		Logb:     false,
		Whiteadd: w,
	}
	data, _ := sonic.ConfigDefault.MarshalIndent(&defaultset, "", "    ")
	utils.WriteFile("settings.json", &data, 0644)
	fmt.Print("Settings.json has been setup. \n")
}

func resync(name *string) {
	_, st := polysync.Databases.Load(*name)
	if !st {
		fmt.Print("There appears to be no databases previous synced...\n")
		return
	} else {
		value, err := utils.ParseJSONFile("databases/" + *name + "/database.json")
		if err != nil {
			fmt.Println("Error unmarshalling Database JSON:", err)
			return
		}
		polysync.Databases.Store(*name, value.Bytes())
		fmt.Print("Resync has been successful!\n")
		value = nil
	}
}

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
