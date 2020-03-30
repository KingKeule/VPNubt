package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
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
	Metric  int
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
	routes := make([]IPv4Routing, 0)
	routeLinesFound := false
	routeLinesOffset := -4 // distance from line matched "IPv4-Routentabelle" to first line with table data
	routeLineSplit := regexp.MustCompile(`\s{2,}`)
	for scanner.Scan() {
		ucl := scanner.Text()

		if 0 == routeLinesOffset {

			if strings.Contains(ucl, "=") {
				break
			}

			split := routeLineSplit.Split(ucl, -1)
			//fmt.Printf("%#v\n", split)

			metric, _ := strconv.Atoi(split[5])

			routes = append(routes,
				IPv4Routing{
					Target:  net.ParseIP(split[1]),
					Mask:    net.ParseIP(split[2]),
					Gateway: net.ParseIP(split[3]),
					Iface:   net.ParseIP(split[4]),
					Metric:  metric})

		} else if !routeLinesFound {
			if strings.Contains(ucl, "IPv4-Routentabelle") {
				routeLinesFound = true
				routeLinesOffset++
			}
		} else {
			routeLinesOffset++
		}
	}

	//fmt.Printf("%s", routes[0])

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
