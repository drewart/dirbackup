package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// Configuration
var (
	logFilePath      = "backup.log"
	concurrencyLevel = 4 // Adjust parallelism here
)

// Logger
var logger *log.Logger

func initLogger() {
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		os.Exit(1)
	}
	logger = log.New(file, "", log.LstdFlags)
}

// shouldCopyFile checks if a file should be copied based on size and modification time.
func shouldCopyFile(srcPath, dstPath string) (bool, error) {
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return false, fmt.Errorf("error getting source file info: %v", err)
	}

	dstInfo, err := os.Stat(dstPath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil // Destination file does not exist
		}
		return false, fmt.Errorf("error getting destination file info: %v", err)
	}

	// Compare file size and modification time
	if srcInfo.Size() != dstInfo.Size() || !srcInfo.ModTime().Equal(dstInfo.ModTime()) {
		return true, nil
	}

	return false, nil
}

// copyFile copies a file from srcPath to dstPath.
func copyFile(srcPath, dstPath string) error {
	dstDir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return fmt.Errorf("error creating destination directory: %v", err)
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("error opening source file: %v", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("error creating destination file: %v", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("error copying file: %v", err)
	}

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("error getting source file info: %v", err)
	}
	if err := os.Chtimes(dstPath, srcInfo.ModTime(), srcInfo.ModTime()); err != nil {
		return fmt.Errorf("error setting modification time: %v", err)
	}

	logger.Println("Copied:", srcPath, "->", dstPath)
	fmt.Println("Copied:", srcPath, "->", dstPath)
	return nil
}

// deleteExtraFiles removes files in the target that don't exist in the source.
func deleteExtraFiles(srcDir, dstDir string) error {
	return filepath.WalkDir(dstDir, func(dstPath string, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %v", dstPath, err)
		}

		// Determine the corresponding source file path
		relPath, err := filepath.Rel(dstDir, dstPath)
		if err != nil {
			return fmt.Errorf("error getting relative path: %v", err)
		}
		srcPath := filepath.Join(srcDir, relPath)

		// If the file or directory does not exist in the source, delete it
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			if d.IsDir() {
				err = os.RemoveAll(dstPath) // Remove directory
			} else {
				err = os.Remove(dstPath) // Remove file
			}
			if err != nil {
				logger.Println("Error deleting:", dstPath, err)
			} else {
				logger.Println("Deleted:", dstPath)
				fmt.Println("Deleted:", dstPath)
			}
		}
		return nil
	})
}

// scanAndBackup scans files and processes them using a worker pool for parallelism.
func scanAndBackup(srcDir, dstDir string, deleteOldFiles bool, dryRun bool) error {
	initLogger()
	defer func() {
		fmt.Println("Backup completed.")
		logger.Println("Backup completed.")
	}()

	fileChan := make(chan string, concurrencyLevel)
	dirChan := make(chan string, concurrencyLevel)
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < concurrencyLevel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for srcPath := range fileChan {
				relPath, err := filepath.Rel(srcDir, srcPath)
				if err != nil {
					logger.Println("Error getting relative path:", err)
					continue
				}
				dstPath := filepath.Join(dstDir, relPath)

				shouldCopy, err := shouldCopyFile(srcPath, dstPath)
				if err != nil {
					logger.Println("Error checking file:", err)
					continue
				}

				if shouldCopy {
					if dryRun {
						logger.Println("[Dry Run] File would be copied:", srcPath, "->", dstPath)
						fmt.Println("[Dry Run] File would be copied:", srcPath, "->", dstPath)
					} else {
						if err := copyFile(srcPath, dstPath); err != nil {
							logger.Println("Error copying file:", err)
						}
					}
				} else {
					logger.Println("File is already up-to-date:", srcPath)
				}
			}
		}()
	}

	// Worker to create empty directories
	wg.Add(1)
	go func() {
		defer wg.Done()
		for dirPath := range dirChan {
			dstPath := filepath.Join(dstDir, dirPath)
			if dryRun {
				logger.Println("[Dry Run] Would create directory:", dstPath)
				fmt.Println("[Dry Run] Would create directory:", dstPath)
			} else {
				if err := os.MkdirAll(dstPath, os.ModePerm); err != nil {
					logger.Println("Error creating directory:", err)
				} else {
					logger.Println("Created directory:", dstPath)
					fmt.Println("Created directory:", dstPath)
				}
			}
		}
	}()

	// Walk through files and directories
	err := filepath.WalkDir(srcDir, func(srcPath string, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %v", srcPath, err)
		}

		relPath, err := filepath.Rel(srcDir, srcPath)
		if err != nil {
			return fmt.Errorf("error getting relative path: %v", err)
		}

		if d.IsDir() {
			dirChan <- relPath // Send directories to be created
		} else {
			fileChan <- srcPath // Send files to be processed
		}
		return nil
	})

	close(fileChan) // Close channels to signal workers to finish
	close(dirChan)
	wg.Wait() // Wait for all workers to complete

	// Delete old files if enabled
	if deleteOldFiles && !dryRun {
		err = deleteExtraFiles(srcDir, dstDir)
	}

	return err
}

func main() {
	var srcDir, dstDir string
	var dryRun bool
	var deleteOldFiles bool
	flag.StringVar(&srcDir, "source", "", "source")
	flag.StringVar(&dstDir, "target", "", "target")
	flag.BoolVar(&dryRun, "dry-run", true, "dry-run")
	flag.BoolVar(&deleteOldFiles, "del-old", false, "delete old files in target")

	flag.Parse()

	err := scanAndBackup(srcDir, dstDir, deleteOldFiles, dryRun)
	if err != nil {
		fmt.Println("Error:", err)
		logger.Println("Error:", err)
	}
}
