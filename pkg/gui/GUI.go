package gui

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"github.com/KingKeule/VPNubt/img"
	"github.com/KingKeule/VPNubt/pkg/config"
)

const appname = "VPNubt"
const version = "v2.0"
const gitHubLink = "https://github.com/KingKeule/VPNubt"
const configFileName = "VPNubt.config"

// Initialization of the GUI
func InitGUI() {

	// hide the windows console window
	// showWindowsConsole(false)

	// ---------------- App/window configuration ----------------
	// Initialize our new fyne interface application.
	app := app.NewWithID("github.com/KingKeule/VPNubt")

	config.Prefs = app.Preferences()
	config.InitAfterFyneApp()

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

	window.SetContent(wireguard())
	window.Resize(fyne.NewSize(300, window.Canvas().Size().Height))

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
// func showWindowsConsole(show bool) {
// 	getConsoleWindow := syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleWindow")
// 	if getConsoleWindow.Find() != nil {
// 		return
// 	}

// 	showWindow := syscall.NewLazyDLL("user32.dll").NewProc("ShowWindow")
// 	if showWindow.Find() != nil {
// 		return
// 	}

// 	hwnd, _, _ := getConsoleWindow.Call()
// 	if hwnd == 0 {
// 		return
// 	}

// 	if show {
// 		showWindow.Call(hwnd, syscall.SW_RESTORE)
// 		log.Println("Windows console window is displayed")
// 	} else {
// 		showWindow.Call(hwnd, syscall.SW_HIDE)
// 		log.Println("Windows console window is hided")
// 	}
// }

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
