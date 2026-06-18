package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var path = flag.String("t", "tunnels.json", "Enter the tunnels json file path")

type Tunnel struct {
	Name         string `json:"name"`
	ListenerAddr string `json:"listen"`
	TargetAddr   string `json:"target"`
}

func run() error {
	file, err := os.Open(*path)
	if err != nil {
		return err
	}
	defer file.Close()

	var tunnels []Tunnel
	if err := json.NewDecoder(file).Decode(&tunnels); err != nil {
		return err
	}

	for _, tunnel := range tunnels {
		t := tunnel

		list, err := net.Listen("tcp", t.ListenerAddr)
		if err != nil {
			log.Printf("Listener error for %s (%s): %v", t.Name, t.ListenerAddr, err)
			continue
		}
		log.Printf("Tunnel [%s] running: %s -> %s", t.Name, t.ListenerAddr, t.TargetAddr)

		go func() {
			defer list.Close()
			for {
				listConn, err := list.Accept()
				if err != nil {
					log.Printf("[%s] Listener connection err: %v", t.Name, err)
					continue
				}
				go func() {
					defer listConn.Close()

					clientIP := listConn.RemoteAddr().String()
					log.Printf("[%s] New connection from %s", t.Name, clientIP)

					targetConn, err := net.Dial("tcp", t.TargetAddr)
					if err != nil {
						log.Printf("[%s] Target connection err: %v", t.Name, err)
						return
					}
					defer targetConn.Close()

					done := make(chan struct{})
					var once sync.Once

					go func() {
						io.Copy(targetConn, listConn)
						once.Do(func() { close(done) })
					}()
					go func() {
						io.Copy(listConn, targetConn)
						once.Do(func() { close(done) })
					}()
					<-done

					log.Printf("[%s] Connection closed for %s", t.Name, clientIP)
				}()
			}
		}()
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	return nil
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
