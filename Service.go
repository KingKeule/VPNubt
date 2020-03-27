package main

import (
	"log"
	"net"
	"time"

	"github.com/tatsushid/go-fastping"
)

func ping(addr string) error {

	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", addr /*inputIP.Text*/)

	if err != nil {
		log.Println(err)
		return err
	}

	p.AddIPAddr(ra)
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		log.Printf("IP Addr: %s receive, RTT: %v\n", addr.String(), rtt)
	}

	p.OnIdle = func() {
		log.Println("finish")
	}

	err = p.Run()
	if err != nil {
		log.Println(err)
	}
	return nil
}
