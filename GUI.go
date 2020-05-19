package main

import (
	"log"
	"net"
	"strconv"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
)

var screenWidth = 280
var screenHight = 400 // not really used because the minimum height is desired
var containerHight = 70

//InitGUI design the GUI of the appp
func InitGUI() {

	// Initialize our new fyne interface application.
	app := app.New()

	// Initialize our new fyne interface application.
	w := app.NewWindow("  VPNubt v0.9")

	// center the windows on the screen
	w.CenterOnScreen()

	// do not allow to resize the window
	w.SetFixedSize(true)

	// Set a sane default for the window size.
	// w.Resize(fyne.NewSize(screenWidth, screenHight))

	// ---------------- Container Configuration ----------------
	// set default values for IP and Port from global config
	defaultConf := getDefaultConf()

	// destiantion IP
	inputdstIP := widget.NewEntry()
	inputdstIP.SetPlaceHolder(defaultConf.dstIP.String())

	// destination port
	inputdstPort := widget.NewEntry()
	inputdstPort.SetPlaceHolder(strconv.Itoa(defaultConf.dstPort))

	// network device
	selectNetDevice := widget.NewSelect(getNetworkInterfaces(), func(selected string) {})

	// create form layout
	widgetdstIPForm := widget.NewFormItem("IP of Server :", inputdstIP)
	widgetdstPortForm := widget.NewFormItem("UDP Port :", inputdstPort)
	widgetNetDevieForm := widget.NewFormItem("    Interface :", selectNetDevice)
	widgetGroupConf := widget.NewGroup("Configuration", widget.NewForm(widgetdstIPForm, widgetdstPortForm, widgetNetDevieForm))

	// ---------------- Container Ping ----------------
	widgetPingStatus := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	buttonPing := widget.NewButton("Ping Server", func() {
		selecteddstIP := net.ParseIP(inputdstIP.Text)
		if (selecteddstIP == nil) || (selecteddstIP.IsUnspecified()) {
			log.Println("The IP address is not correct. Please set a valid IP adress.")
		} else {
			log.Println("Start pinging server (IP: " + selecteddstIP.String() + ")")
			recieved, err := ping(selecteddstIP.String())
			if err != nil || !recieved {
				widgetPingStatus.SetText("NOK")
			} else {
				widgetPingStatus.SetText("OK")
			}
		}
	})

	widgetGroupPing := widget.NewGroup("Ping", fyne.NewContainerWithLayout(layout.NewGridLayout(2),
		buttonPing, widgetPingStatus))

	// ---------------- Container Service Command ----------------
	widgetTunnelServiceStat := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})

	// create a channel to communicate the stop command to capture & forwart thread
	stopThreadChannel := make(chan bool)

	// boolean value to differentiate whether to start the tunneling service or not
	serviceRunning := false

	// button for start or stop tunneling service
	// currently no easy way to change the name of the button. alternatively 2 buttons could be set.
	buttonTunnelServiceStat := widget.NewButton("Start / Stop", func() {
		port, err := strconv.Atoi(inputdstPort.Text)
		selecteddstIP := net.ParseIP(inputdstIP.Text)
		if (selecteddstIP == nil) || (selecteddstIP.IsUnspecified()) {
			log.Println("The IP address is not correct. Please set a valid IP adress.")
		} else if (err != nil) || (port < 1) || (port > 65535) {
			log.Println("The UDP port is not correct. Please set a UDP valid port (1-65535).")
		} else if selectNetDevice.Selected == "" {
			log.Println("The selection of the network device is not set. Please select network device.")
		} else {
			if !serviceRunning {
				log.Println("Starting udp broadcast tunneling service")
				widgetTunnelServiceStat.SetText("Running")
				go capturePackets(stopThreadChannel, selectNetDevice.Selected, selecteddstIP, port)
				serviceRunning = true
			} else {
				log.Println("Stopping udp broadcast tunneling service")
				stopThreadChannel <- true // Send stop signal to channel.
				widgetTunnelServiceStat.SetText("Stopped")
				serviceRunning = false
			}
		}
	})

	widgetGroupTunnelService := widget.NewGroup("Tunnelling Service", fyne.NewContainerWithLayout(layout.NewGridLayout(2),
		buttonTunnelServiceStat, widgetTunnelServiceStat))

	// ---------------- Container complete ----------------
	containerAll := fyne.NewContainerWithLayout(layout.NewGridLayout(1),
		widgetGroupConf,
		//FixedGridLayout required cause GridLayout divides the space equally for all cells
		fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize(screenWidth, containerHight)),
			widgetGroupPing,
			widgetGroupTunnelService))

	w.SetContent(containerAll)

	// define and add the menu to the window
	w.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("Tool",
			fyne.NewMenuItem("Reset configuration", func() {
				defaultConf := getDefaultConf()
				//TODO find an better way for update the variables and move menu ahead
				inputdstIP.SetText(defaultConf.dstIP.String())
				inputdstPort.SetText(strconv.Itoa(defaultConf.dstPort))
				//TODO not possible to go back to default choice
				// selectProtocolTpye.SetSelected("Select one")
				widgetPingStatus.SetText("")
				widgetTunnelServiceStat.SetText("")
			})),
		fyne.NewMenu("Preload Configuration",
			fyne.NewMenuItem("Warcraft 3", func() {
				w3Conf := getWar3Conf()
				//TODO find an better way for update the variables and move menu ahead
				inputdstPort.SetText(strconv.Itoa(w3Conf.dstPort))
			})),
		fyne.NewMenu("Help",
			fyne.NewMenuItem("Show Log", func() {
				//TODO Integrate Logviewer
			}),
			fyne.NewMenuItem("About", func() {
				//TODO Integrate PopUp
			}),
		)))

	// Show all of our set content and run the gui.
	w.ShowAndRun()
}
