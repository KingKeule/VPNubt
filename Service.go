package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"os/exec"

	"github.com/tatsushid/go-fastping"
)

type config struct {
	network net.IP // the address to reach default gateway
}

type IPv4Routing struct {
	Target  net.IP
	Mask    net.IP
	Gateway net.IP
	Iface   net.IP
	Metric  uint
}

func initService() {
	cmd := exec.Command("route", "print")
	r, w := io.Pipe()
	cmd.Stdout = w
	go func() {
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()

	scanner := bufio.NewScanner(r)
	//routes := make([]IPv4Routing, 10)
	for scanner.Scan() {
		ucl := scanner.Text()
		target := strings.Contains(ucl, "IPv4-Routentabelle")
		if target {
			fmt.Println(ucl)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return
	}
}

func ping(addr string) (bool, error) {

	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", addr /*inputIP.Text*/)
	var recieved bool = false

	if err != nil {
		log.Println(err)
		return recieved, err
	}

	p.AddIPAddr(ra)
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		recieved = true
		log.Printf("IP Addr: %s receive, RTT: %v\n", addr.String(), rtt)
	}

	p.OnIdle = func() {
		log.Println("Ping job finished")
	}

	err = p.Run()
	if err != nil {
		log.Println(err)
	}

	return recieved, nil
}
