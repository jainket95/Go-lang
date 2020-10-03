package main

import (
	"encoding/json"
	"fmt"
	"io"

	// "io"00
	"io/ioutil"
	// "log"
	"os"
	// "path/filepath"
)

type AbsolutePaths struct {
	AbsolutePaths []Path `json:"absolutePaths"`
}

type Path struct {
	Path string `json:"path"`
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	fmt.Println("Starting the app")

	jsonFile, err := os.Open("users.json")

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successfully Opened users.json")

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var aplha AbsolutePaths

	json.Unmarshal(byteValue, &aplha)

	os.Remove("backup")

	err = os.Mkdir("backup", 0755)
	check(err)

	// var caches = filepath.Dir("caches")
	// fmt.Println(caches)

	for i := 0; i < len(aplha.AbsolutePaths); i++ {
		fmt.Println("Directories Path: " + aplha.AbsolutePaths[i].Path)
		path := "backup"

		sourceFileStat, err := os.Stat(aplha.AbsolutePaths[i].Path)
		if err != nil {
			check(err)
		}

		if !sourceFileStat.Mode().IsRegular() {
			fmt.Errorf("%s is not a regular file", aplha.AbsolutePaths[i].Path)
		}

		source, err := os.Open(aplha.AbsolutePaths[i].Path)
		if err != nil {
			check(err)
		}
		defer source.Close()

		destination, err := os.Create(path)
		if err != nil {
			check(err)
		}
		defer destination.Close()
		nBytes, err := io.Copy(destination, source)
		check(err)
		fmt.Println(nBytes)
	}
}
