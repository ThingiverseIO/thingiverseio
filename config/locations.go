package config

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

const (
	CFG_FILE_NAME         = "thingiverseio.conf"
	CFG_GLOBAL_PATH_LINUX = "/etc/thingiverse.io/"
	CFG_USER_PATH_LINUX   = ".thingiverse.io/"
)

func CfgFileCwd() (dir string) {
	dir, _ = os.Getwd()
	dir = filepath.Join(dir, CFG_FILE_NAME)
	return
}

func CfgFileCwdPresent() bool {
	_, err := os.Stat(CfgFileCwd())
	return err == nil
}

func CfgFileGlobal() (dir string) {
	switch runtime.GOOS {
	case "linux":
		dir = CFG_GLOBAL_PATH_LINUX
	}
	dir = filepath.Join(dir, CFG_FILE_NAME)
	return
}

func CfgFileGlobalPresent() bool {
	_, err := os.Stat(CfgFileGlobal())
	return err == nil
}

func CfgFileUser() (dir string) {
	usr, _ := user.Current()
	switch runtime.GOOS {
	case "linux":
		dir = filepath.Join(usr.HomeDir, CFG_USER_PATH_LINUX)
	}
	dir = filepath.Join(dir, CFG_FILE_NAME)
	return
}

func CfgFileUserPresent() bool {
	_, err := os.Stat(CfgFileUser())
	return err == nil
}
