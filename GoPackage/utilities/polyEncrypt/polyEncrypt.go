package polySecurity

import (
	"crypto/rc4"
	"fmt"
	"os"

	"github.com/Jeffail/gabs/v2"
	utils "github.com/JewishLewish/PolygonDB/GoPackage/utilities/polyFuncs"
	polygon "github.com/JewishLewish/PolygonDB/GoPackage/utilities/polyStructs"
	polysync "github.com/JewishLewish/PolygonDB/GoPackage/utilities/polySync"
	"github.com/bytedance/sonic"
)

func Encrypt(target *string) {
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

func Decrypt(target *string) {
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

func indent(input *[]byte) {
	var output interface{}
	sonic.ConfigDefault.Unmarshal(*input, &output)
	*input, _ = sonic.ConfigDefault.MarshalIndent(&output, "", "    ")
}
