package main

import "github.com/cloudtask/libtools/gounits/system"

import (
	"log"
	"os"
)

func main() {

	jobworker, err := NewJobWorker()
	if err != nil {
		log.Printf("jobworker error, %s\n", err.Error())
		os.Exit(system.ErrorExitCode(err))
	}

	defer func() {
		exitCode := 0
		if err := jobworker.Stop(); err != nil {
			log.Printf("jobworker stop error, %s\n", err.Error())
			exitCode = system.ErrorExitCode(err)
		}
		os.Exit(exitCode)
	}()

	if err = jobworker.Startup(); err != nil {
		log.Printf("jobworker startup error, %s\n", err.Error())
		os.Exit(system.ErrorExitCode(err))
	}
	system.InitSignal(nil)
}
