package internal

import (
	"encoding/json"
	"log"
	"os"
	"os/user"
	"path"
)

type (
	singleProjectJson map[string]int64
	lockFileJson      map[string]singleProjectJson
)

type Lockfile struct {
	files []string
	JSON  lockFileJson
}

func NewLockfile(files []string) Lockfile {
	return Lockfile{
		files: files,
	}
}

// Loads existing lock information generates it for the first time.
func (l *Lockfile) Bootstrap() {
	lockfilePath, err := l.getLockfilePath()
	if err != nil {
		log.Fatal(err)
	}

	if !FileExists(lockfilePath) {
		l.generateLockfile(true)
	}

	currentLockFile, err := os.ReadFile(lockfilePath)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(currentLockFile, &l.JSON)
	if err != nil {
		log.Fatal(err)
	}
}

// Returns the lock information for the current project.
func (l *Lockfile) GetCurrentProject() singleProjectJson {
	cwd, _ := os.Getwd()
	return l.JSON[cwd]
}

// Update timestamps for files in current project.
func (l *Lockfile) UpdateTimestampsForFiles(files []string) error {
	lockfileMap, err := l.prepareMap(files)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	l.JSON[cwd] = lockfileMap
	for f := range l.JSON[cwd] {
		l.JSON[cwd][f] = lockfileMap[f]
	}

	err = l.generateLockfile(false)
	if err != nil {
		return err
	}

	return nil
}

// Generate the lockfile file, or update it with new contents.
func (l *Lockfile) generateLockfile(initialLockfile bool) error {
	contents := l.JSON
	if initialLockfile {
		lockfileMap, err := l.prepareMap(l.files)
		if err != nil {
			return err
		}

		cwd, _ := os.Getwd()
		contents = lockFileJson{cwd: lockfileMap}
	}

	jsonString, err := json.MarshalIndent(contents, "", "  ")
	if err != nil {
		return err
	}

	writeCh := make(chan error)
	go l.writeLockfileRoutine(jsonString, writeCh)

	if err = <-writeCh; err != nil {
		return err
	}

	return nil
}

// Prepares the map used to populate individual project files.
func (l *Lockfile) prepareMap(files []string) (singleProjectJson, error) {
	lockfileMapCh := make(chan Ref[singleProjectJson])
	go l.getFileModifiedMapRoutine(files, lockfileMapCh)

	lockfileRef := <-lockfileMapCh

	if lockfileRef.Error() != nil {
		return nil, lockfileRef.Error()
	}

	return lockfileRef.Value(), nil
}

// Go routine used to dispatch file mtime checks in the background.
func (l *Lockfile) getFileModifiedMapRoutine(files []string, ch chan Ref[singleProjectJson]) {
	lockfileMap := make(singleProjectJson)

	for _, f := range files {
		fo, err := os.Stat(f)

		if err != nil {
			ch <- NewRef[singleProjectJson](nil, err)
		}

		lockfileMap[f] = fo.ModTime().Unix()
	}

	ch <- NewRef(lockfileMap, nil)
}

// Writes the lockfile into the filesystem.
func (l *Lockfile) writeLockfileRoutine(contents []byte, ch chan error) {
	gokePath, err := l.getLockfilePath()
	if err != nil {
		ch <- err
		return
	}

	if err = os.WriteFile(gokePath, contents, 0644); err != nil {
		ch <- err
		return
	}

	ch <- nil
}

// Returns the location of the lockfile in the system.
func (l *Lockfile) getLockfilePath() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}

	return path.Join(user.HomeDir, ".goke"), nil
}
