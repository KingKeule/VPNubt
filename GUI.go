package main

import (
	"log"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
)

var screenWidth = 280
var screenHight = 400

//InitGUI design the GUI of the appp
func InitGUI() {

	// Initialize our new fyne interface application.
	app := app.New()

	// Initialize our new fyne interface application.
	w := app.NewWindow("VPN packet tunneler")

	// Set a sane default for the window size.
	w.Resize(fyne.NewSize(screenWidth, screenHight))

	// ---------------- Container Configuration ----------------
	inputIP := widget.NewEntry()
	inputIP.SetPlaceHolder("192.168.XXX.XXX")

	inputPort := widget.NewEntry()
	inputPort.SetPlaceHolder("6112")

	widgetGroupConf := widget.NewGroup("Configuration", &widget.Form{
		// Items: []*widget.FormItem{{"IP of Server", inputIP}, {"Port of Server", inputPort}, {"", widget.NewLabel("")}}, // ToDo: Remove when spacer is working
		Items: []*widget.FormItem{{"IP of Server", inputIP}, {"Port of Server", inputPort}},
	})

	// ---------------- Container Ping ----------------
	buttonPing := widget.NewButton("Ping Server", func() {

		log.Println("Start pinging server (" + inputIP.Text + ")")
		err := ping(inputIP.Text)
		if err != nil {
			log.Println("NOK")
			log.Println(err)
		} else {
			log.Println("OK")
		}
	})

	widgetPingStatus := widget.NewLabel("OK")
	widgetPingStatus.TextStyle = fyne.TextStyle{Bold: false}
	widgetPingStatus.Alignment = fyne.TextAlignCenter

	widgetGroupPing := widget.NewGroup("Ping", fyne.NewContainerWithLayout(layout.NewGridLayout(2),
		buttonPing, widgetPingStatus))

	widgetTunnelServiceStat := widget.NewLabel("Not running")
	widgetTunnelServiceStat.TextStyle = fyne.TextStyle{Bold: false}
	widgetTunnelServiceStat.Alignment = fyne.TextAlignCenter

	// ---------------- Container Service Command ----------------
	buttonTunnelServiceStat := widget.NewButton("Start", func() {
		log.Println("Start tunneling service")
		//widgetTunnelServiceStat = widget.NewLabel("Running") //ToDo : Change Status of widget
	})
	//buttonTunnelServiceStat.Resize(fyne.NewSize(2 00, 200)) // ToDo: Resize not working. Change size -> smaler

	widgetGroupTunnelService := widget.NewGroup("Tunnelling Service", fyne.NewContainerWithLayout(layout.NewGridLayout(2),
		buttonTunnelServiceStat, widgetTunnelServiceStat))

	// ToDo: Spacer espcially the size of the spacer are not really working
	// spacer := layout.NewSpacer()
	// spacer.Resize(fyne.NewSize(500, 500))

	// ---------------- Container complete ----------------
	// Add all the buttons in to a three column grid layout inside a container.
	containerAll := fyne.NewContainerWithLayout(layout.NewGridLayout(1),
		widgetGroupConf,
		// spacer,
		widgetGroupPing,
		widgetGroupTunnelService)
	// rect)

	w.SetContent(containerAll)

	// Show all of our set content and run the gui.
	w.ShowAndRun()
}
