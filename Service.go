package main

import (
	"bufio"
	"encoding/hex"
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

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/tatsushid/go-fastping"
)

type config struct {
	defaultIface  net.IP // the interface to reach default gateway
	pcapIface     pcap.Interface
	pcapHandle    *pcap.Handle
	pcapPortUDP   uint
	VPNGateway    string // TODO from 'arp -a' to net.HardwareAddr
	VPNHostTarget net.IP
	VPNHostSource net.IP
}

var currentConfig config

type IPv4Routing struct {
	Target  net.IP
	Mask    net.IP
	Gateway net.IP
	Iface   net.IP
	Metric  int
}

func AnyPcapAddress(these []pcap.InterfaceAddress, other net.IP) bool {
	for _, cur := range these {
		if cur.IP.Equal(other) {
			return true
		}
	}
	return false
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

	// routes already sorted by metric, so take first
	// TODO consider dynamic changes
	currentConfig.defaultIface = routes[0].Iface

	allEthDevs, err := pcap.FindAllDevs()

	if err != nil {
		panic(err)
	}

	for _, ethDev := range allEthDevs {
		if AnyPcapAddress(ethDev.Addresses, currentConfig.defaultIface) {
			currentConfig.pcapIface = ethDev
			break
		}
	}

	fmt.Println("Using '", currentConfig.pcapIface.Description, "' to Capture Broadcast Packets")

	handle, err := pcap.OpenLive(currentConfig.pcapIface.Name, 65536, false, pcap.BlockForever)
	if err != nil {
		panic(err)
	}

	currentConfig.pcapHandle = handle
	currentConfig.VPNGateway = "c8:0e:14:30:fd:d7" // TODO from arp
}

func startServicePcap() {

	fmt.Println("Configuring Broadcast Packet Capture on UDP Port", currentConfig.pcapPortUDP)

	filter := fmt.Sprintf("host %s and ip broadcast and udp port %d",
		currentConfig.defaultIface.String(),
		currentConfig.pcapPortUDP)

	fmt.Println("Setting Filter '" + filter + "'")

	if err := currentConfig.pcapHandle.SetBPFFilter(filter); err != nil {
		panic(err)
	}

	packetSource := gopacket.NewPacketSource(currentConfig.pcapHandle, currentConfig.pcapHandle.LinkType())
	for packet := range packetSource.Packets() {

		var thisSrcMAC net.HardwareAddr
		var thisDstMAC net.HardwareAddr
		var thisSrcIP net.IP
		var thisDstIP net.IP
		var thisSrcPort layers.UDPPort
		var thisDstPort layers.UDPPort

		ethLayer := packet.Layer(layers.LayerTypeEthernet)
		if ethLayer == nil {
			panic("Layer Ethernet is nil")
		}
		eth, _ := ethLayer.(*layers.Ethernet)
		thisSrcMAC = eth.SrcMAC
		thisDstMAC = eth.DstMAC

		ip4Layer := packet.Layer(layers.LayerTypeIPv4)
		if ip4Layer == nil {
			panic("Layer IPv4 is nil")
		}
		ip4, _ := ip4Layer.(*layers.IPv4)
		thisSrcIP = ip4.SrcIP
		thisDstIP = ip4.DstIP

		udpLayer := packet.Layer(layers.LayerTypeUDP)
		if udpLayer == nil {
			panic("Layer UDP is nil")
		}
		udp, _ := udpLayer.(*layers.UDP)
		thisSrcPort = udp.SrcPort
		thisDstPort = udp.DstPort

		fmt.Printf("%s %d:%d -> %s %d:%d\n", thisSrcMAC, thisSrcIP, thisSrcPort, thisDstMAC, thisDstIP, thisDstPort)

		// redirect paket

		gwMACDecoded, err := hex.DecodeString(strings.ReplaceAll(currentConfig.VPNGateway, ":", ""))
		if err != nil {
			log.Fatal(err)
		}

		eth.DstMAC = gwMACDecoded
		ip4.DstIP = currentConfig.VPNHostTarget

		fmt.Printf("Redirecting...\n")
		fmt.Printf("%s %d:%d -> %s %d:%d\n", thisSrcMAC, thisSrcIP, thisSrcPort, currentConfig.VPNGateway, ip4.DstIP, thisDstPort)

		buf := gopacket.NewSerializeBuffer()
		opts := gopacket.SerializeOptions{}

		gopacket.SerializeLayers(buf, opts,
			eth,
			ip4,
			udp,
			gopacket.Payload(packet.ApplicationLayer().Payload()))

		packetData := buf.Bytes()

		currentConfig.pcapHandle.WritePacketData(packetData)
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
