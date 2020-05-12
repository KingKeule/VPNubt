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
	inputPort := widget.NewEntry()
	inputPort.SetPlaceHolder(strconv.Itoa(defaultConf.dstPort))

	// network device
	labelNetDevice := widget.NewLabelWithStyle("    Interface :", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	selectNetDevice := widget.NewSelect(getNetworkDevices(), func(selected string) {})
	containerNetDevice := fyne.NewContainerWithLayout(layout.NewGridLayout(2),
		labelNetDevice, selectNetDevice)

	widgetGroupConf := widget.NewGroup("Configuration", &widget.Form{
		// Items: []*widget.FormItem{{"IP of Server", inputIP}, {"Port of Server", inputPort}, {"", widget.NewLabel("")}}, // ToDo: Remove when spacer is working
		Items: []*widget.FormItem{{"IP of Server :", inputdstIP}, {"UDP Port :", inputPort}},
	}, containerNetDevice)

	// ---------------- Container Ping ----------------
	widgetPingStatus := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	buttonPing := widget.NewButton("Ping Server", func() {
		if (inputdstIP.Text == "") || (inputdstIP.Text == "0.0.0.0") {
			log.Println("The IP address is not correct. Please set a valid IP adress.")
		} else {
			log.Println("Start pinging server (IP: " + inputdstIP.Text + ")")
			recieved, err := ping(inputdstIP.Text)
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
	widgetTunnelServiceStat := widget.NewLabelWithStyle("Not running", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})

	buttonTunnelServiceStat := widget.NewButton("Start", func() {
		port, err := strconv.Atoi(inputPort.Text)
		if (inputdstIP.Text == "") || (inputdstIP.Text == "0.0.0.0") {
			log.Println("The IP address is not correct. Please set a valid IP adress.")
		} else if (err != nil) || (port < 1) || (port > 65535) {
			log.Println("The UDP port is not correct. Please set a UDP valid port (1-65535).")
		} else if selectNetDevice.Selected == "" {
			log.Println("The selection of the network device is not set. Please select network device.")
		} else {
			log.Println("Start tunneling service")
			// capture packets and forward them to the set IP address
			capturePackets(selectNetDevice.Selected, net.ParseIP(inputdstIP.Text), port)
			//widgetTunnelServiceStat = widget.NewLabel("Running") //ToDo : Change Status of widget
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

	w.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("Tool",
			fyne.NewMenuItem("Reset configuration", func() {
				defaultConf := getDefaultConf()
				//TODO find an better way for update the variables and move menu ahead
				inputdstIP.SetText(defaultConf.dstIP.String())
				inputPort.SetText(strconv.Itoa(defaultConf.dstPort))
				//TODO not possible to go back to default choice
				// selectProtocolTpye.SetSelected("Select one")
				widgetPingStatus.SetText("")
			}),
			fyne.NewMenuItem("Import configuration", func() {})),
		fyne.NewMenu("Preload Configuration",
			fyne.NewMenuItem("Warcraft 3", func() {
				w3Conf := getWar3Conf()
				//TODO find an better way for update the variables and move menu ahead
				inputPort.SetText(strconv.Itoa(w3Conf.dstPort))
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
