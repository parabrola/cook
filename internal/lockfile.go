package internal

import (
	"encoding/json"
	"os/user"
	"path"
)

type (
	singleProjectJson map[string]int64
	lockFileJson      map[string]singleProjectJson
)

type Lockfile struct {
	files   []string
	JSON    lockFileJson
	options Options
	fs      FileSystem
}

func NewLockfile(files []string, opts *Options, fs FileSystem) Lockfile {
	return Lockfile{
		files:   files,
		options: *opts,
		fs:      fs,
	}
}

func (l *Lockfile) Bootstrap() error {
	lockfilePath, err := l.getLockfilePath()
	if err != nil {
		return err
	}

	if !l.fs.FileExists(lockfilePath) {
		if err := l.generateLockfile(true); err != nil {
			return err
		}
	}

	currentLockFile, err := l.fs.ReadFile(lockfilePath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(currentLockFile, &l.JSON); err != nil {
		return err
	}

	return nil
}

// Returns the lock information for the current project.
func (l *Lockfile) GetCurrentProject() singleProjectJson {
	cwd, _ := l.fs.Getwd()
	return l.JSON[cwd]
}

// Update timestamps for files in current project.
func (l *Lockfile) UpdateTimestampsForFiles(files []string) error {
	lockfileMap, err := l.prepareMap(files)
	if err != nil {
		return err
	}

	cwd, err := l.fs.Getwd()
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

		cwd, _ := l.fs.Getwd()
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
		fo, err := l.fs.Stat(f)

		if err != nil {
			ch <- NewRef[singleProjectJson](nil, err)
			return
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

	if err = l.fs.WriteFile(gokePath, contents, 0644); err != nil {
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
