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
func CaptureAndForwardPacket(stopThreadChannel chan bool, pktCntChannel chan bool, dstIP net.IP, dstPort int) {
	//selection, whether to start or stop the service when the stop signal comes

	// create an udp socket
	log.Printf("Listen on UDP port: %d", dstPort)
	conn, err := net.ListenPacket("udp", ":"+strconv.Itoa(dstPort))
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
				rlen, srcAddr, err := conn.ReadFrom(buf[:])
				if err != nil && !strings.Contains(err.Error(), "timeout") {
					log.Println(err)
				}

				// forward packet
				if rlen > 0 {
					log.Printf("UDP packet (payload: %d bytes) was received from %s and is being forwarded to %s:%d", rlen, srcAddr, dstIP.String(), dstPort)
					srcAddr := strings.Split(srcAddr.String(), ":")
					srcPort, err := strconv.Atoi(srcAddr[1])
					if err != nil {
						log.Println(err)
						return
					}
					// this method must not be multi-threading, otherwise there will be duplicate stack creation which will not work
					forwardPacket(conn, dstIP, dstPort, srcPort, buf[0:rlen], pktCntChannel)
				}
			}
		}
	}()
}

func forwardPacket(conn net.PacketConn, dstIP net.IP, dstPort int, srcPort int, data []byte, pktCntChannel chan bool) {
	//normally the destination port is also the source port.
	//however, in order to generically forward all packets correctly and not to manipulate them, there is a check here.
	//if the ports are different a new socket must be created otherwise the existing one is used.
	if srcPort != dstPort {
		conn, err := net.ListenPacket("udp", ":"+strconv.Itoa(srcPort))
		if err != nil {
			log.Println(err)
			return
		}

		defer conn.Close()
	}

	dstAddr := net.UDPAddr{
		Port: dstPort,
		IP:   dstIP,
	}

	_, err := conn.WriteTo(data, &dstAddr)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("UDP packet was successfully forwarded to %s:%d", dstIP, dstPort)
		pktCntChannel <- true // increase the counter of forwarded packet
	}
}
