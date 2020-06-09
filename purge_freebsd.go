package main

import (
	"time"
	"os"
	"syscall"
	"io/ioutil"
	"path/filepath"
)

func purge() {
	files, err := ioutil.ReadDir(outputPath)
	if err != nil {
		log_Errorf("Error reading output dir: %+v", err)
		return
	}

	cutoffTime := time.Now().Add(time.Hour * 24 * -time.Duration(recordingKeepDays))
	for _, f := range files {
		fullPath := filepath.Join(outputPath, f.Name())
		info, err := os.Stat(fullPath)
		if err != nil {
			log_Errorf("Error reading file (%s): %+v", fullPath, err)
			continue
		}

		stat_t := info.Sys().(*syscall.Stat_t)
		created := time.Unix(int64(stat_t.Ctimespec.Sec), int64(stat_t.Ctimespec.Nsec))
		
		if created.Before(cutoffTime) {
			err := os.Remove(fullPath)
			if err != nil {
				log_Errorf("Error deleting file (%s): %+v", fullPath, err)
			}
		}
	}
}
