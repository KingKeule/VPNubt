package service

import (
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/tatsushid/go-fastping"
)

// checks whether the given ip host address can be pinged
func Ping(addr string) (bool, error) {

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
		log.Fatal(err)
	}

	return recieved, nil
}

// capture all udp broadcast packets on given port and forward them to the given destination ip address
// via the StopThreadChannel this function receives the information from the GUI to be stopped
func CaptureAndForwardPacket(stopThreadChannel chan bool, dstIP net.IP, dstPort int) {
	//selection, whether to start or stop the service when the stop signal comes

	// resolve the address for given port
	srcAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(dstPort))
	if err != nil {
		log.Println(err)
		return
	}

	// create an udp socket
	log.Printf("Listen on UDP port: %d", dstPort)
	conn, err := net.ListenUDP("udp", srcAddr)
	if err != nil {
		log.Println(err)
		return
	}

	go func() {
		defer func() {
			conn.Close()
			log.Println("UDP socket closed")
		}()

		buf := make([]byte, 512)

		for {
			select {
			case <-stopThreadChannel:
				log.Println("UDP broadcast tunneling service stopped")
				return

			default:
				// Set a deadline for reading. Read operation will fail if no data is received after deadline.
				timeout := 3 * time.Second
				err = conn.SetReadDeadline(time.Now().Add(timeout))
				if err != nil {
					log.Println(err)
					return
				}
				//log.Printf("UDP connection timeout is set to %s. Timeouts are not logged.", timeout)

				//read packet on given udp port
				rlen, _, err := conn.ReadFromUDP(buf[:])
				if err != nil && !strings.Contains(err.Error(), "timeout") {
					log.Println(err)
				}

				// forward packet
				if rlen > 0 {
					log.Printf("UDP packet (payload: %d bytes) was received and is being forwarded to %s:%d", rlen, dstIP.String(), dstPort)

					dstAddr := net.UDPAddr{
						Port: dstPort,
						IP:   net.ParseIP(dstIP.String()),
					}

					_, err := conn.WriteTo(buf[0:rlen], &dstAddr)
					if err != nil {
						log.Println(err)
					} else {
						log.Printf("UDP packet was successfully forwarded to %s:%d", dstIP.String(), dstPort)
					}
				}
			}
		}
	}()
}
