package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

// writePid writes current process ID into a file
func writePid(file string) error {
	if file == "" {
		return nil
	}
	log.Debugf("writing PID to %s", file)
	pid := strconv.Itoa(os.Getpid())
	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(pid)
	return err
}

// removePid removes file containing the process ID
func removePid(file string) error {
	if file == "" {
		return nil
	}
	if _, err := os.Stat(file); err == nil {
		log.Debugf("removing PID file %s", file)
		return os.Remove(file)
	}
	return nil
}

// waitPid check if another instance is running and eventually wait for it to terminate (with a timeout)
func waitPid(pidFile string, waitTimeout int) error {
	if pidFile == "" {
		return nil
	}

	waitTime := 100 * time.Millisecond // initial backoff time for waiting
	maxBackoffTime := 1000             // maximum time in millisecond to add between each waiting runs

	timeout := false
	maxTime := time.Now().Add(time.Second * time.Duration(waitTimeout))

	for {
		if _, err := os.Stat(pidFile); err == nil {
			// PID file exists
			log.Debugf("waiting %s before checking PID again", waitTime)
			time.Sleep(waitTime)
			waitTime = waitTime + time.Duration(rand.Intn(maxBackoffTime))*time.Millisecond
		} else {
			// file does not exist
			break
		}
		if time.Now().After(maxTime) {
			timeout = true
			break
		}
	}

	if timeout {
		return fmt.Errorf("timeout waiting for PID file")
	}

	return nil
}
