package main

import (
	"log"
	"net"
	"os/exec"
	"strconv"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
)

var screenWidth = 280
var screenHight = 400 // not really used because the minimum height is desired
var containerHight = 70

const version = "v0.9"
const gitHubLink = "https://github.com/KingKeule/VPN-broadcast-tunneler"

//InitGUI design the GUI of the appp
func InitGUI() {

	// Initialize our new fyne interface application.
	app := app.New()

	// set the theme for the app. Default is dark theme
	//app.Settings().SetTheme(theme.LightTheme())

	// Initialize our new fyne interface application.
	w := app.NewWindow("  VPNubt " + version)

	// indicates that closing this main window should exit the app
	w.SetMaster()

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
		if !checkIPAdress(selecteddstIP, w) {
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

		if !checkIPAdress(selecteddstIP, w) {
		} else if !checkPort(err, port, w) {
		} else if !checkNetDevice(selectNetDevice.Selected, w) {
		} else if !serviceRunning {
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
		fyne.NewMenu("Game selection",
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
				// windows command to open the browser with the given link
				exec.Command("rundll32", "url.dll,FileProtocolHandler", gitHubLink).Start()
			}),
		)))

	// Show all of our set content and run the gui.
	w.ShowAndRun()
}

// checks whether the given string is a valid IP address
// if not then an error message is displayed on the given window
func checkIPAdress(ip net.IP, window fyne.Window) bool {
	if (ip == nil) || (ip.IsUnspecified()) {
		ipWarnText1 := "The entered IP address is not correct."
		ipWarnText2 := "Please set a valid IP adress."
		log.Println(ipWarnText1 + " " + ipWarnText2)
		dialog.ShowInformation("", ipWarnText1+"\n"+ipWarnText2, window)
		return false
	}
	return true
}

// checks whether the given string is a valid port
// if not then an error message is displayed on the given window
func checkPort(err error, port int, window fyne.Window) bool {
	if (err != nil) || (port < 1) || (port > 65535) {
		portWarnText1 := "The entered UDP port is not correct."
		portWarnText2 := "Please set a UDP valid port (1-65535)."
		log.Println(portWarnText1 + " " + portWarnText2)
		dialog.ShowInformation("", portWarnText1+"\n"+portWarnText2, window)
		return false
	}
	return true
}

// checks whether a network device is selected
// if not then an error message is displayed on the given window
func checkNetDevice(netDevice string, window fyne.Window) bool {
	if netDevice == "" {
		netDevWarnText1 := "The network device is not set."
		netDevWarnText2 := "Please select network device."
		log.Println(netDevWarnText1 + " " + netDevWarnText2)
		dialog.ShowInformation("", netDevWarnText1+"\n"+netDevWarnText2, window)
		return false
	}
	return true
}
