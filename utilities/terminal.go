package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/bytedance/sonic"
)

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
	WriteFile(conpath, &cinput, 0644)

	datapath := path + "/database.json"
	dinput := []byte("{\n\t\"Example\": \"Hello world\"\n}")
	WriteFile(datapath, &dinput, 0644)

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
	WriteFile("settings.json", &data, 0644)
	fmt.Print("Settings.json has been setup. \n")
}
