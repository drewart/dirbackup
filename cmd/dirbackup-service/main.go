package main

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Config struct {
	SourceDir           string   `json:"source_dir"`
	TargetDir           string   `json:"target_dir"`
	Folders             []string `json:"folders"`
	FolderDelayDuration string   `json:"folder-delay-duration"`
	SleepDuration       string   `json:"sleep-duration"`
	DirBackupPath       string   `json:"dirbackup-path"`
}

/*func dameon(config Config) {

	ticker := time.NewTicker(time.Duration(config.DelaySeconds) * time.Second)
	workers := make(chan bool, 1)
	death := make(chan os.Signal, 1)
	signal.Notify(death, os.Interrupt, os.Kill)
	folderIndex := 0
	for {
		select {
		case <-ticker.C:
			log.Println("Scheduled task is triggered.")
			if folderIndex >= len(config.Folders) {
				folderIndex = 0
				log.Println("Starting Over Sleeping")
				time.Sleep(time.Duration(config.DelaySeconds))
			}
			folderIndex++
			source := config.SourceDir + "/" + config.Folders[folderIndex]
			target := config.BackupDir + "/" + config.Folders[folderIndex]
			go runWorker(source, target, workers)
		case <-workers:
			log.Println("Scheduled task is completed.")
			// can't return, it needs to be continue running
		case <-death:
			//do any clean up you need and return
			log.Println("service killed")
		}
	}
}*/

func dameonRun(config Config) {

	folderDelay, err := time.ParseDuration(config.FolderDelayDuration)
	if err != nil {
		log.Fatal(err)
	}
	sleepDelay, err := time.ParseDuration(config.SleepDuration)
	if err != nil {
		log.Fatal(err)
	}

	folderIndex := 0
	for {
		if folderIndex >= len(config.Folders) {
			folderIndex = 0
			log.Printf("Starting Over Sleeping %s\n", config.SleepDuration)
			time.Sleep(sleepDelay)
		}
		source := config.SourceDir + "/" + config.Folders[folderIndex]
		target := config.TargetDir + "/" + config.Folders[folderIndex]
		runWorkerBackup(config, source, target)
		log.Printf("Folder Sleep: %s\n", config.FolderDelayDuration)
		time.Sleep(folderDelay)
		folderIndex++
	}
}

func runWorkerBackup(config Config, source, target string) {
	command := []string{config.DirBackupPath, "-source", source, "-target", target, "-dry-run=false", "-del-old=true"}
	log.Println(strings.Join(command, " "))
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Run()
}

/*func runWorker(source, target string, workers chan bool) {
	cmd := exec.Command("dirbackup", "-source", source, "-target", target, "-dry-run", "false", "-del-old", "true")
	cmd.Run()
	// do the work
	workers <- true
}*/

func loadConfig(configFile string) (Config, error) {
	var config Config
	file, err := os.OpenFile(configFile, os.O_RDONLY, 0644)
	if err != nil {
		return config, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func main() {
	configFile := os.Getenv("BACKUP_CONFIG")
	if configFile == "" {
		curDir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		configFile = curDir + "/bkservice.json"
	}
	// overwrite BACKUP_CONFIG with param
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	log.Printf("Loadding config %s\n:", configFile)
	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}
	if config.SleepDuration == "" {
		config.SleepDuration = "30m"
	}
	if config.FolderDelayDuration == "" {
		config.FolderDelayDuration = "5m"
	}
	if config.DirBackupPath == "" {
		config.DirBackupPath = "dirbackup"
	}

	log.Println("Starting Dameon")
	dameonRun(config)
}
