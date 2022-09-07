package main

import (
	"encoding/json"
	"log"
	"os"
)

type Lockfile struct {
	Files []string
}

func (l *Lockfile) GenerateLockfile() {
	lockfileMapCh := make(chan map[string]int64)
	go l.getFileModifiedMap(lockfileMapCh)
	lockfileMap := <-lockfileMapCh

	jsonString, err := json.MarshalIndent(lockfileMap, "", "  ")
	if err != nil {
		log.Fatalln(err)
	}

	writeCh := make(chan error)
	go l.writeLockfile(jsonString, writeCh)

	if err = <-writeCh; err != nil {
		log.Fatalln(err)
	}
}

func (l *Lockfile) getFileModifiedMap(ch chan map[string]int64) {
	lockfileMap := make(map[string]int64)

	for _, f := range l.Files {
		fo, err := os.Stat(f)

		if err != nil {
			log.Fatalln(err)
		}

		lockfileMap[f] = fo.ModTime().Unix()
	}

	ch <- lockfileMap
}

func (l *Lockfile) writeLockfile(contents []byte, ch chan error) {
	err := os.WriteFile("goke.lock", contents, 0644)

	if err != nil {
		ch <- err
		return
	}

	ch <- nil
}
