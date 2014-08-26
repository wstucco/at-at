package at_at

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"runtime"
)

const (
	StartingPort = 16000
	DotFolder    = ".at-at"
)

var currentPort = StartingPort

func Run() {
	hosts := scan(home())

	router := NewRouter(hosts)
	router.Run()
}

func scan(folder string) HostList {
	root := path.Join(folder, DotFolder)
	files, error := ioutil.ReadDir(root)
	links := make(HostList, 0)

	if error == nil {
		for _, f := range files {
			if isSymLink(f) {
				link := path.Join(root, f.Name())
				links[f.Name()] = NewHost(f.Name(), link, nextPort())
			}
		}
	}

	return links
}

func home() string {

	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}

	return true
}

func isSymLink(file os.FileInfo) bool {
	return file.Mode()&os.ModeSymlink == os.ModeSymlink
}

func nextPort() int {

	// possible infinite loop
	// FIXME
	for {
		if conn, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", currentPort)); err == nil {
			conn.Close()
			break
		}
		currentPort++
	}

	return currentPort
}
