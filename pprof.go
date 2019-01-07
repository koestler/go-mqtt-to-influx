package main

import (
	"log"
	"os"
	"runtime"
	"runtime/pprof"
)

func runCpuProfile(fileName string) (started bool) {
	if fileName == "" {
		return false
	}

	f, err := os.Create(fileName)
	if err != nil {
		log.Printf("pprof: could not open file for CPU profile: %s", err)
		return false
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Printf("pprof: could not start CPU profile: %s", err)
		return false
	}
	log.Print("pprof: started CPU profile, save data to: %s", fileName)

	return true
}

func writeMemProfile(fileName string) {
	if fileName == "" {
		return
	}

	f, err := os.Create(fileName)
	if err != nil {
		log.Printf("pprof: could not create memory profile: %s", err)
	}
	runtime.GC() // get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Printf("pprof: could not write memory profile: %s", err)
	}
	log.Print("pprof: wrote memory profile to %s", fileName)
	_ = f.Close()
}
