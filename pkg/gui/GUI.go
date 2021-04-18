package gui

import (
	"encoding/json"
	"io/ioutil"
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
	"github.com/KingKeule/VPNubt/img"
	"github.com/KingKeule/VPNubt/pkg/config"
	"github.com/KingKeule/VPNubt/pkg/service"
)

var screenWidth = 280
var screenHight = 400 // not really used because the minimum height is desired

const appname = "VPNubt"
const version = "v1.2"
const gitHubLink = "https://github.com/KingKeule/VPNubt"
const configFileName = "VPNubt.config"

//Initialization of the GUI
func InitGUI() {

	// hide the windows console window
	showWindowsConsole(false)

	// ---------------- App/window configuration ----------------
	// Initialize our new fyne interface application.
	app := app.New()

	// set the theme for the app. Default is dark theme
	app.Settings().SetTheme(theme.DarkTheme())

	// set the logo of the application
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
	conf := checkForConfig()

	// destination IP
	inputdstIP := widget.NewEntry()
	inputdstIP.Text = conf.DstIP

	// destination port
	inputdstPort := widget.NewEntry()
	inputdstPort.Text = strconv.Itoa(conf.SrcPort)

	// create form layout
	widgetdstIPForm := widget.NewFormItem("IP of Server :", inputdstIP)
	widgetdstPortForm := widget.NewFormItem("UDP Port :", inputdstPort)
	widgetGroupConf := widget.NewGroup("Configuration", widget.NewForm(widgetdstIPForm, widgetdstPortForm))

	// ---------------- Container Ping ----------------
	widgetPingStatus := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	buttonPing := widget.NewButton("Ping Server", func() {
		selecteddstIP := net.ParseIP(inputdstIP.Text)
		if !checkIPAddress(selecteddstIP, window) {
		} else {
			log.Println("Start pinging server (IP: " + selecteddstIP.String() + ")")
			recieved, err := service.Ping(selecteddstIP.String())
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

		if !checkIPAddress(selecteddstIP, window) {
		} else if !checkPort(err, port, window) {
		} else if !serviceRunning {
			log.Println("Starting udp broadcast tunneling service")
			widgetTunnelServiceStat.SetText("Running")
			service.CaptureAndForwardPacket(stopThreadChannel, selecteddstIP, port)
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
				defaultConf := config.GetDefaultConf()
				//TODO find an better way for update the variables and move menu ahead
				inputdstIP.SetText(defaultConf.DstIP)
				inputdstPort.SetText(strconv.Itoa(defaultConf.SrcPort))
				widgetPingStatus.SetText("")
				widgetTunnelServiceStat.SetText("")
				log.Println("Reset of all input and status fields")
			}),
			fyne.NewMenuItem("Save configuration", func() {
				port, err := strconv.Atoi(inputdstPort.Text)
				selecteddstIP := net.ParseIP(inputdstIP.Text)
				if checkIPAddress(selecteddstIP, window) {
					if checkPort(err, port, window) {
						writeConfigToFile(inputdstIP.Text, port, window)
					}
				}
			})),
		fyne.NewMenu("Game selection",
			fyne.NewMenuItem("Warcraft 3", func() {
				w3Conf := config.GetWar3Conf()
				//TODO find an better way for update the variables and move menu ahead
				inputdstPort.SetText(strconv.Itoa(w3Conf.SrcPort))
				log.Println("Set udp port (" + strconv.Itoa(w3Conf.SrcPort) + ") for selected game: Warcraft 3")
			}),
			fyne.NewMenuItem("CoD - UO", func() {
				coDUOConf := config.GetCoDUOConf()
				//TODO find an better way for update the variables and move menu ahead
				inputdstPort.SetText(strconv.Itoa(coDUOConf.SrcPort))
				log.Println("Set udp port (" + strconv.Itoa(coDUOConf.SrcPort) + ") for selected game: Call of Duty - United Offensive")
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
func checkIPAddress(ip net.IP, window fyne.Window) bool {
	if (ip == nil) || (ip.IsUnspecified()) {
		ipWarnText1 := "The entered IP address is not correct."
		ipWarnText2 := "Please set a valid IP address."
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

// write the current given IP address and udp port to file
func writeConfigToFile(inputdstIP string, inputsrcPort int, window fyne.Window) {
	log.Println("Try to write actual IP address and udp port to a config file in the workspace")

	actConfig := config.Config{DstIP: inputdstIP, SrcPort: inputsrcPort}

	//marshal (pretty) to json structure
	jsonConfigData, err := json.MarshalIndent(actConfig, "", "  ")

	err = ioutil.WriteFile(configFileName, jsonConfigData, 0644)
	if err != nil {
		log.Println(err)
		dialog.ShowInformation("", "Error while saving.\n See log for more details", window)
	} else {
		log.Println("Configuration file (VPNubt.config) was saved successfully.")
		dialog.ShowInformation("", "Configuration file (VPNubt.config)\n was saved successfully ", window)
	}
}

func checkForConfig() *config.Config {
	configFileBytes, err := ioutil.ReadFile(configFileName)
	if err != nil {
		log.Printf("No config file (%s) was found in the working directory. Using no config.", configFileName)
		return config.GetDefaultConf()
	}

	log.Printf("Config file (%s) was found in the working directory and will be imported.", configFileName)
	var actConfig config.Config
	err = json.Unmarshal([]byte(configFileBytes), &actConfig)
	if err != nil {
		log.Printf("Error while marshalling config file. %s.", err)
		return config.GetDefaultConf()
	} else {
		return &actConfig
	}
}

// Logo return the visual logo for the window icon
// Icon source: https://icons8.de/icons/set/tunnel
func Logo() fyne.Resource {
	icon := &fyne.StaticResource{
		StaticName:    "icons8-tunnel-24.png",
		StaticContent: img.IconBytes}
	return icon
}
