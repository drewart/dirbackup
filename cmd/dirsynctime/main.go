package main

import (
	"fmt"
	"log"
	"os"
)

func main() {

	dir := "./"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, e := range entries {
		st, err := os.Stat(e.Name())
		if err != nil {
			log.Printf("Error getting file info: %v\n", err)
		}
		fmt.Printf("Name: %s, Size: %d, Mode: %v, ModTime: %v Name: %s\n", e.Name(), st.Size(), st.Mode(), st.ModTime(), e.Name())
	}
}
