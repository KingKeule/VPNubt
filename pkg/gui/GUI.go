package gui

import (
	"encoding/json"
	"image/color"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/KingKeule/VPNubt/img"
	"github.com/KingKeule/VPNubt/pkg/config"
	"github.com/KingKeule/VPNubt/pkg/service"
)

var screenWidth float32 = 280
var menuHeight float32 = 26

var fwdPktCnt int = 0

const appname = "VPNubt"
const version = "v2.1"
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

	// ---------------- event channels  ----------------
	// create a channel to communicate the stop command to capture & forwart thread
	stopThreadChannel := make(chan bool)

	// create a channel to notify a new forwarded packet or to reset the forwarded packet counter
	pktCntChannel := make(chan bool)

	// ---------------- Container Configuration ----------------
	// set default values for IP and Port from global config
	conf := checkForConfig()

	// destination IP
	inputDstIP := widget.NewEntry()
	inputDstIP.Text = conf.DstIP

	// destination port
	inputDstPort := widget.NewEntry()
	inputDstPort.Text = strconv.Itoa(conf.DstPort)

	// create form layout
	widgetDstIPForm := widget.NewFormItem("IP of Server:", inputDstIP)
	widgetDstPortForm := widget.NewFormItem("UDP Port:", inputDstPort)
	widgetGroupConf := widget.NewForm(widgetDstIPForm, widgetDstPortForm)

	// ---------------- Container Ping ----------------
	widgetPingStatus := widget.NewLabelWithStyle("---", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	buttonPing := widget.NewButton("Ping Server", func() {})
	buttonPing.OnTapped = func() {
		selectedDstIP := net.ParseIP(inputDstIP.Text)
		if !checkIPAddress(selectedDstIP, window) {
			return
		}
		log.Println("Ping IP:", selectedDstIP.String())
		pong := make(chan bool)
		go service.Ping(selectedDstIP.String(), pong)
		go func() {
			buttonPing.Disable()
			defer buttonPing.Enable()
			timeout := time.NewTimer(1100 * time.Millisecond)
			defer timeout.Stop()
			select {
			case p := <-pong:
				if p {
					widgetPingStatus.SetText("OK")
				} else {
					widgetPingStatus.SetText("NOK")
				}
			case <-timeout.C:
				widgetPingStatus.SetText("NOK")
			}
		}()
	}

	widgetGroupPing := container.NewGridWithColumns(2, buttonPing, widgetPingStatus)

	// ---------------- Container status ----------------
	widgetServiceStatLabel := widget.NewLabelWithStyle("Service:", fyne.TextAlignTrailing, fyne.TextStyle{Bold: false})
	widgetServiceStatValue := widget.NewLabelWithStyle("Stopped", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	widgetPktCntLabel := widget.NewLabelWithStyle("Fwd packets:", fyne.TextAlignTrailing, fyne.TextStyle{Bold: false})
	WidgetPktCntValue := widget.NewLabelWithStyle("0", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	widgetGroupStatus := fyne.NewContainerWithLayout(layout.NewGridLayout(2),
		widgetServiceStatLabel, widgetServiceStatValue, widgetPktCntLabel, WidgetPktCntValue)

	// ---------------- Container Service Command ----------------
	// boolean value to differentiate whether to start the tunneling service or not
	serviceRunning := false

	// since the button is transparent and cannot have any color, a color layer is integrated
	bgLayerRed := canvas.NewRectangle(color.NRGBA{R: 200, G: 0, B: 0, A: 200})
	bgLayerGreen := canvas.NewRectangle(color.NRGBA{R: 0, G: 200, B: 0, A: 200})

	// button for start or stop tunneling service
	// currently no easy way to change the name of the button. alternatively 2 buttons could be set.
	buttonTunnelServiceStat := widget.NewButton("Start", nil)
	buttonTunnelServiceStat.OnTapped = func() {
		dstPort, err := strconv.Atoi(inputDstPort.Text)
		dstIP := net.ParseIP(inputDstIP.Text)

		if !checkIPAddress(dstIP, window) {
		} else if !checkPort(err, dstPort, window) {
		} else if !serviceRunning {
			log.Println("Starting UDP broadcast tunneling service")
			widgetServiceStatValue.SetText("Running")
			buttonTunnelServiceStat.SetText("Stop")
			bgLayerGreen.Hide()
			service.CaptureAndForwardPacket(stopThreadChannel, pktCntChannel, dstIP, dstPort)
			serviceRunning = true
		} else {
			log.Println("Stopping UDP broadcast tunneling service")
			stopThreadChannel <- true // Send stop signal to channel.
			widgetServiceStatValue.SetText("Stopped")
			buttonTunnelServiceStat.SetText("Start")
			bgLayerGreen.Show()
			serviceRunning = false
		}
	}

	containerServiceCmd := container.New(layout.NewMaxLayout(), container.New(layout.NewPaddedLayout(), bgLayerRed, bgLayerGreen), buttonTunnelServiceStat)

	// ---------------- Container complete ----------------
	containerAll := fyne.NewContainerWithLayout(layout.NewVBoxLayout(),
		NewGroupCustom("Configuration"),
		widgetGroupConf,
		NewGroupCustom("Ping"),
		widgetGroupPing,
		NewGroupCustom("Tunneling Service"),
		containerServiceCmd,
		NewGroupCustom("Status"),
		widgetGroupStatus)
	window.SetContent(containerAll)

	// Resize only in width due the menu width and take the actual height of the window
	window.Resize(fyne.NewSize(screenWidth, window.Canvas().Size().Height+menuHeight))

	// ---------------- Menu ----------------
	// define and add the menu to the window
	window.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("Tool",
			fyne.NewMenuItem("Reset all values", func() {
				defaultConf := config.GetDefaultConf()
				//TODO find an better way for update the variables and move menu ahead
				inputDstIP.SetText(defaultConf.DstIP)
				inputDstPort.SetText(strconv.Itoa(defaultConf.DstPort))
				widgetPingStatus.SetText("")
				widgetServiceStatValue.SetText("")
				pktCntChannel <- false // reset the counter and display of forwarded packets
				log.Println("Reset of all input and status fields")
			}),
			fyne.NewMenuItem("Save configuration", func() {
				dstIP := net.ParseIP(inputDstIP.Text)
				if checkIPAddress(dstIP, window) {
					dstPort, err := strconv.Atoi(inputDstPort.Text)
					if checkPort(err, dstPort, window) {
						writeConfigToFile(inputDstIP.Text, dstPort, window)
					}
				}
			})),
		fyne.NewMenu("Game selection",
			fyne.NewMenuItem("Warcraft 3", func() {
				w3Conf := config.GetWar3Conf()
				//TODO find an better way for update the variables and move menu ahead
				inputDstPort.SetText(strconv.Itoa(w3Conf.DstPort))
				log.Println("Set UDP port (" + strconv.Itoa(w3Conf.DstPort) + ") for selected game: Warcraft 3")
			}),
			fyne.NewMenuItem("CoD - UO", func() {
				coDUOConf := config.GetCoDUOConf()
				//TODO find an better way for update the variables and move menu ahead
				inputDstPort.SetText(strconv.Itoa(coDUOConf.DstPort))
				log.Println("Set UDP port (" + strconv.Itoa(coDUOConf.DstPort) + ") for selected game: Call of Duty - United Offensive")
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

	//  start the gui update mechanism e.g. for the packet counter widget
	go updateGui(WidgetPktCntValue, pktCntChannel)

	// Show all of our set content and run the gui.
	window.ShowAndRun()
}

// gui update mechanism
func updateGui(widgetPktCntValue *widget.Label, pktCntChannel chan bool) {
	for {
		select {

		case pktCntChannelValue := <-pktCntChannel:
			if pktCntChannelValue == true {
				fwdPktCnt += 1
				widgetPktCntValue.SetText(strconv.Itoa(fwdPktCnt))
			} else {
				fwdPktCnt = 0
				widgetPktCntValue.SetText("")
			}
		}
	}
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
	log.Printf("Try to write actual IP address and udp port to a configuration file (%s) in the workspace", configFileName)

	actConfig := config.Config{DstIP: inputdstIP, DstPort: inputsrcPort}

	//marshal (pretty) to json structure
	jsonConfigData, err := json.MarshalIndent(actConfig, "", "  ")

	err = ioutil.WriteFile(configFileName, jsonConfigData, 0644)
	if err != nil {
		log.Println(err)
		dialog.ShowInformation("", "Error while saving.\n See log for more details", window)
	} else {
		log.Printf("Configuration file (%s) was saved successfully.", configFileName)
		dialog.ShowInformation("", "Configuration file ("+configFileName+")\n was saved successfully", window)
	}
}

func checkForConfig() *config.Config {
	log.Printf("Try to read the configuration file (%s) from workspace", configFileName)

	configFileBytes, err := ioutil.ReadFile(configFileName)
	if err != nil {
		log.Printf("No configuration file (%s) was found in the working directory. Using no config", configFileName)
		return config.GetDefaultConf()
	}

	log.Printf("Configuration file (%s) was found in the working directory and will be imported", configFileName)
	var actConfig config.Config
	err = json.Unmarshal([]byte(configFileBytes), &actConfig)
	if err != nil {
		log.Printf("Error while marshalling configuration file. %s", err)
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
