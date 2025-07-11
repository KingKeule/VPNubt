package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
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

	window.SetContent(wireguard(window))
	window.Resize(fyne.NewSize(300, window.Canvas().Size().Height))

	// Show all of our set content and run the gui.
	window.ShowAndRun()
}

// Logo return the visual logo for the window icon
// Icon source: https://icons8.de/icons/set/tunnel
func Logo() fyne.Resource {
	icon := &fyne.StaticResource{
		StaticName:    "icons8-tunnel-24.png",
		StaticContent: img.IconBytes}
	return icon
}
