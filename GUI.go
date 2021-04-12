package main

import (
	_ "embed"
	"log"
	"net"
	"os/exec"
	"strconv"
	"syscall"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

var screenWidth = 280
var screenHight = 400 // not really used because the minimum height is desired

const appname = "VPNubt"
const version = "v1.2"
const gitHubLink = "https://github.com/KingKeule/VPNubt"

//InitGUI design the GUI of the appp
func InitGUI() {

	// hide the windows console window
	showWindowsConsole(false)

	// ---------------- App/window configuration ----------------
	// Initialize our new fyne interface application.
	app := app.New()

	// set the theme for the app. Default is dark theme
	app.Settings().SetTheme(theme.DarkTheme())

	// set the logo iof the application
	app.SetIcon(Logo())

	// Initialize our new fyne interface application.
	window := app.NewWindow(" " + appname + " " + version)

	// indicates that closing this main window should exit the app
	window.SetMaster()

	// center the windows on the screen
	window.CenterOnScreen()

	// do not allow to resize the window
	window.SetFixedSize(true)

	// ---------------- Container Configuration ----------------
	// set default values for IP and Port from global config
	defaultConf := getDefaultConf()

	// destination IP
	inputdstIP := widget.NewEntry()
	inputdstIP.SetPlaceHolder(defaultConf.dstIP.String())

	// destination port
	inputdstPort := widget.NewEntry()
	inputdstPort.SetPlaceHolder(strconv.Itoa(defaultConf.dstPort))

	// create form layout
	widgetdstIPForm := widget.NewFormItem("IP of Server :", inputdstIP)
	widgetdstPortForm := widget.NewFormItem("UDP Port :", inputdstPort)
	widgetGroupConf := widget.NewGroup("Configuration", widget.NewForm(widgetdstIPForm, widgetdstPortForm))

	// ---------------- Container Ping ----------------
	widgetPingStatus := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	buttonPing := widget.NewButton("Ping Server", func() {
		selecteddstIP := net.ParseIP(inputdstIP.Text)
		if !checkIPAdress(selecteddstIP, window) {
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

		if !checkIPAdress(selecteddstIP, window) {
		} else if !checkPort(err, port, window) {
		} else if !serviceRunning {
			log.Println("Starting udp broadcast tunneling service")
			widgetTunnelServiceStat.SetText("Running")
			captureAndForwardPacket(stopThreadChannel, selecteddstIP, port)
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
	containerAll := fyne.NewContainerWithLayout(layout.NewVBoxLayout(),
		widgetGroupConf,
		widgetGroupPing,
		widgetGroupTunnelService)
	window.SetContent(containerAll)

	// // Resize only in width due the menü width and take the actual height of the window
	window.Resize(fyne.NewSize(screenWidth, window.Canvas().Size().Height))

	// ---------------- Menu ----------------
	// define and add the menu to the window
	window.SetMainMenu(fyne.NewMainMenu(
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
				log.Println("Reset of all input and status fields")
			})),
		fyne.NewMenu("Game selection",
			fyne.NewMenuItem("Warcraft 3", func() {
				w3Conf := getWar3Conf()
				//TODO find an better way for update the variables and move menu ahead
				inputdstPort.SetText(strconv.Itoa(w3Conf.dstPort))
				log.Println("Set udp port (" + strconv.Itoa(w3Conf.dstPort) + ") for selected game: Warcraft 3")
			}),
			fyne.NewMenuItem("CoD - UO", func() {
				coDUOConf := getCoDUOConf()
				//TODO find an better way for update the variables and move menu ahead
				inputdstPort.SetText(strconv.Itoa(coDUOConf.dstPort))
				log.Println("Set udp port (" + strconv.Itoa(coDUOConf.dstPort) + ") for selected game: Call of Duty - United Offensive")
			})),
		fyne.NewMenu("Help",
			fyne.NewMenuItem("Show Log", func() {
				showWindowsConsole(true)
			}),
			fyne.NewMenuItem("About", func() {
				// windows command to open the browser with the given link
				exec.Command("rundll32", "url.dll,FileProtocolHandler", gitHubLink).Start()
				log.Println("Open github site from the project")
			}),
		)))

	// Show all of our set content and run the gui.
	window.ShowAndRun()
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
		portWarnText2 := "Please set a valid UDP port (1-65535)."
		log.Println(portWarnText1 + " " + portWarnText2)
		dialog.ShowInformation("", portWarnText1+"\n"+portWarnText2, window)
		return false
	}
	return true
}

// https://stackoverflow.com/questions/23743217/printing-output-to-a-command-window-when-golang-application-is-compiled-with-ld/23744350
// https://forum.golangbridge.org/t/no-println-output-with-go-build-ldflags-h-windowsgui/7633/6
// this functions open the windows standard console window
func showWindowsConsole(show bool) {
	getConsoleWindow := syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleWindow")
	if getConsoleWindow.Find() != nil {
		return
	}

	showWindow := syscall.NewLazyDLL("user32.dll").NewProc("ShowWindow")
	if showWindow.Find() != nil {
		return
	}

	hwnd, _, _ := getConsoleWindow.Call()
	if hwnd == 0 {
		return
	}

	if show {
		showWindow.Call(hwnd, syscall.SW_RESTORE)
		log.Println("Windows console window is displayed")
	} else {
		showWindow.Call(hwnd, syscall.SW_HIDE)
		log.Println("Windows console window is hided")
	}
}

//go:embed img\\icons8-tunnel-24.png
var iconBytes []byte

// Logo return the visual logo for the window icon
// Icon source: https://icons8.de/icons/set/tunnel
func Logo() fyne.Resource {
	icon := &fyne.StaticResource{
		StaticName:    "icons8-tunnel-24.png",
		StaticContent: iconBytes}
	return icon
}
