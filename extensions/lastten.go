package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	_ = sort("[{'age':5,'name':'A'},{'age':20,'name':'B'},{'age':30,'name':'C'},{'age':40,'name':'D'}]")
}

func say_hello() string {
	fmt.Print("Hello world\n")
	return ("Hello world")
}

func sort(data string) string {
	var jsonData interface{}
	json.Unmarshal([]byte(data), &jsonData)
	fmt.Print(data, "\n")
	return ("Hello world")
}
