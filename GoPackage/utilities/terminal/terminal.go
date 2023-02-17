package terms

import (
	"bufio"
	"crypto/rc4"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/Jeffail/gabs/v2"
	utils "github.com/JewishLewish/PolygonDB/GoPackage/utilities/polyFuncs"
	polygon "github.com/JewishLewish/PolygonDB/GoPackage/utilities/polyStructs"
	polysync "github.com/JewishLewish/PolygonDB/GoPackage/utilities/polySync"
	"github.com/bytedance/sonic"
)

// Terminal
// This is designed for the standalone executable.
// However, datacreate() is used in the Create Function for Go Package
func Terminal() {
	scanner := bufio.NewScanner(os.Stdin)
	locked := false
	var lock string = ""
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

func encrypt(target *string) {
	var jsonData polygon.Config
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
		err = polysync.Datacheck(target, &database)
		if err != nil {
			fmt.Print(err)
			return
		}

		newtext := deep_encrypt(database.Bytes(), []byte(jsonData.Key))
		fmt.Print(string(newtext), "\n")

		jsonData.Enc = true

		output, _ := sonic.ConfigDefault.MarshalIndent(&jsonData, "", "    ")
		utils.WriteFile("databases/"+*target+"/config.json", &output, 0644)
		utils.WriteFile("databases/"+*target+"/database.json", &newtext, 0644)

		fmt.Print("Encryption successful for " + *target + ".\n")
	} else {
		fmt.Print("The following data is already encrypted. Don't encrypt again.\n")
	}
}

func decrypt(target *string) {
	var jsonData polygon.Config
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
		utils.WriteFile("databases/"+*target+"/config.json", &output, 0644)
		utils.WriteFile("databases/"+*target+"/database.json", &newtext, 0644)

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
