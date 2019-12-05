package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	Info   *log.Logger
	Error  *log.Logger // 非常严重的问题
	LogPath    string
	SkynetPath string
	ExeHomeDir string
	Version  = "v1.0"
)

func init() {
	createALogDirectory()
	Info = log.New(io.MultiWriter(openLogFile("/info.log"), os.Stderr),"INFO: ",log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(io.MultiWriter(openLogFile("/error.log"), os.Stderr),"ERROR: ",log.Ldate|log.Ltime|log.Lshortfile)
}

func exe_home_dir(){
	dir, _ := os.Executable()
	exPath := filepath.Dir(dir)
	ExeHomeDir = exPath
}

func UserConf() map[string]interface{} {
	user_conf := make(map[string]interface{})
	bytes, e := ioutil.ReadFile(SkynetPath+"/user.conf")
	if e == nil {
		json.Unmarshal(bytes, &user_conf)
	}
	return user_conf
}

func createALogDirectory()  {
	exe_home_dir()
	if ExeHomeDir == "" {
		os.Exit(0)
	}else {
		if strings.HasSuffix(ExeHomeDir,"/") {
			ExeHomeDir = ExeHomeDir[0:len(ExeHomeDir) -1]
		}
		SkynetPath = ExeHomeDir + "/skynet"
		mkdir := dirMkdir(SkynetPath)
		if mkdir != nil {
			log.Println("create skynet error",mkdir)
			os.Exit(0)
		}else {
			LogPath = SkynetPath + "/logs"
			mkdir := dirMkdir(LogPath)
			if mkdir != nil {
				log.Println("create logs error",mkdir)
				os.Exit(0)
			}
		}
	}
}

func dirMkdir(path string) error {
	_, e := os.Stat(path)
	if e != nil {
		e := os.MkdirAll(path, os.ModePerm)
		if e != nil {
			return e
		}
	}
	return nil
}

func openLogFile(fileName string) *os.File {
	openFile, e := os.OpenFile(LogPath+fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if e != nil {
		fmt.Println(e)
		os.Exit(0)
	}
	return openFile
}