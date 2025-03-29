package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	SourceDir         string `json:"source_dir"`
	BackupDir         string `json:"backup_dir"`
	LastFile          string `json:"last_file"`
	LastTimeStamp     int64  `json:"last_timestamp"`
	LastScanTimeStamp int64  `json:"last_scan_timestamp"`
}

var (
	config         Config
	origTimeStamps int64
	lastScan       int64
)

func main() {

	var configFile string
	flag.StringVar(&configFile, "config", "dirscan.json", "Configuration file")

	flag.Parse()

	config, err := loadConfig(configFile)

	lastScan = config.LastScanTimeStamp
	config.LastScanTimeStamp = time.Now().Unix()
	if err != nil {
		fmt.Print("Error loading config file:", err)
	}
	origTimeStamps = config.LastTimeStamp

	curDir, err := os.Getwd()
	config.SourceDir = curDir
	if err != nil {
		log.Fatal(err)
	}
	err = filepath.WalkDir(curDir, walkFn)
	if err != nil {
		fmt.Println("Error walking directory:", err)
	}
	saveConfig(configFile, config)
}

func loadConfig(configFile string) (Config, error) {
	var config Config
	file, err := os.OpenFile(configFile, os.O_RDONLY, 0644)
	if err != nil {
		return config, err
	}
	defer file.Close()
	configData, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(configData, &config)
	return config, nil
}

func saveConfig(configFile string, config Config) {
	file, err := os.OpenFile(configFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	file.Write(configData)
	fmt.Println(string(configData))
}

func walkFn(path string, info os.DirEntry, err error) error {
	if err != nil {
		fmt.Println("Error accessing path:", path, err)
		return nil // Skip the path and continue walking
	}
	st, err := os.Stat(path)
	if err != nil {
		log.Printf("Error getting file info: %v\n", err)
	}
	ts := st.ModTime().Unix()
	if ts > config.LastTimeStamp {
		config.LastTimeStamp = ts
		config.LastFile = path
	}
	if ts > lastScan {
		fmt.Println("copy:", path, "Mod:", st.ModTime())
	}

	//fmt.Printf("Name: %s, Size: %d, Mode: %v, ModTime: %v Name: %s\n", e.Name(), st.Size(), st.Mode(), st.ModTime(), e.Name())
	return nil
}
