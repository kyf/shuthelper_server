package main

import (
	"encoding/json"
	"flag"
	"github.com/kardianos/service"
	"net"
	"strings"
)

var (
	exit   chan int = make(chan int)
	logger service.Logger
)

type program struct{}

func getLocalAddr() map[string]string {
	conn, err := net.Dial("udp", "www.baidu.com:80")
	if err != nil {
		logger.Error(err)
		return nil
	}
	defer conn.Close()
	addr := strings.Split(conn.LocalAddr().String(), ":")
	return map[string]string{
		"IP":   addr[0],
		"Port": addr[1],
	}
}

func (p *program) Start(s service.Service) error {
	go p.run(s)
	return nil
}

func (p *program) run(s service.Service) {
	localAddr := getLocalAddr()
	if localAddr == nil {
		s.Stop()
		return
	}
	socket, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 6060,
	})
	if err != nil {
		logger.Error(err)
		s.Stop()
		return
	}

	logger.Info("udp server has start ...")
	defer socket.Close()
	out := make([]byte, 1024)
	for {
		select {
		case <-exit:
			return
		default:
			_, addr, err := socket.ReadFromUDP(out)
			if err != nil {
				logger.Error(err)
				continue
			}
			logger.Infof("receive [%v] data : %s \n", addr.IP, string(out))
			resp := map[string]string{
				"ip": localAddr["IP"],
			}
			r, err := json.Marshal(resp)
			if err != nil {
				logger.Error(err)
				continue
			}
			socket.WriteToUDP(r, &net.UDPAddr{
				IP:   addr.IP,
				Port: 6060,
			})
		}
	}
}

func (p *program) Stop(s service.Service) error {
	close(exit)
	return nil
}

func main() {
	controlFlag := flag.String("control", "", "{install|uninstall|start|stop}")
	flag.Parse()
	conf := &service.Config{
		Name:        "shuthelper_server",
		DisplayName: "shuthelper server service",
		Description: "remote host shutdown helper service ...",
	}

	p := &program{}

	s, err := service.New(p, conf)
	if err != nil {
		panic(err)
	}

	logger, err = s.Logger(nil)
	if err != nil {
		panic(err)
	}

	if !strings.EqualFold(*controlFlag, "") {
		err = service.Control(s, *controlFlag)
		if err != nil {
			panic(err)
		}
		return
	}

	s.Run()
}
