package main

import (
	"github.com/kardianos/service"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// service
type program struct {
	exit chan struct{}
}

func (p *program) Start(s service.Service) error {
	p.exit = make(chan struct{})
	go func() {
		run := p.run()
		if run != nil {
			Error.Println(run)
			p.Stop(s)
		}
	}()
	return nil
}
func (p *program) run() error {
	go func() {
		guide := installationGuide()
		if guide != nil {
			Error.Println(guide)
			p.Stop(Skynet_service)
		}
	}()
	go func() {
		goos := runtime.GOOS
		Info.Println("system :", goos)
		if strings.HasPrefix(goos, "win") {
			cmd := exec.Command(`cmd`, `/c`, `start`, `http://127.0.0.1:50000`)
			e := cmd.Start()
			if e != nil {
				Error.Println("win open browser error", e)
			}
		} else if strings.HasPrefix(goos, "darwin") {
			e := exec.Command(`open`, `http://127.0.0.1:50000`).Start()
			if e != nil {
				Error.Println("mac open browser error", e)
			}
		}
		Info.Println("授权地址:","http://127.0.0.1:50000")
	}()
	userConf := UserConf()
	if len(userConf) > 0 {
		port, e := strconv.Atoi(userConf["port"].(string))
		if e != nil {
			return e
		}
		start(userConf["key"].(string), userConf["public_ip"].(string), port, nil)
		Info.Println("已经启动")
	} else {
		Info.Println("未授权")
	}
	return nil
}
func (p *program) Stop(s service.Service) error {
	// Any work in Stop should be quick, usually a few seconds at most.
	Info.Println("stop skynet service")
	close(p.exit)
	return nil
}

var Skynet_service service.Service

func init() {
	options := make(service.KeyValue)
	options["Restart"] = "on-success"
	options["SuccessExitStatus"] = "1 2 8 SIGKILL"
	svcConfig := &service.Config{
		Name:        "skynet",
		DisplayName: "公网IP",
		Description: "公网IP",
		//Dependencies: []string{
		//	"Requires=network.target",
		//	"After=network-online.target syslog.target"},
		Option: options,
	}
	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	} else {
		Skynet_service = s
	}
}