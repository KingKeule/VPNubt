package main

import (
	"log"
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
	w := app.NewWindow("VPN packet tunneler")

	// Set a sane default for the window size.
	// w.Resize(fyne.NewSize(screenWidth, screenHight))

	// ---------------- Container Configuration ----------------
	// set default values for IP and Port from global config
	inputIP := widget.NewEntry()
	inputIP.SetPlaceHolder(dstIP.String())

	inputPort := widget.NewEntry()
	inputPort.SetPlaceHolder(strconv.Itoa(dstPort))

	labelProtocolTpye := widget.NewLabelWithStyle("Protocol type", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	selectProtocolTpye := widget.NewSelect([]string{"UDP", "TCP"}, func(selected string) {
		switch selected {
		case "UDP":
			protocoltpye = "UDP"
		case "TCP":
			protocoltpye = "TCP"
		}
	})
	containerProtocolTpye := fyne.NewContainerWithLayout(layout.NewGridLayout(2),
		labelProtocolTpye, selectProtocolTpye)

	widgetGroupConf := widget.NewGroup("Configuration", &widget.Form{
		// Items: []*widget.FormItem{{"IP of Server", inputIP}, {"Port of Server", inputPort}, {"", widget.NewLabel("")}}, // ToDo: Remove when spacer is working
		Items: []*widget.FormItem{{"IP of Server", inputIP}, {"Port of Server", inputPort}},
	}, containerProtocolTpye)

	// ---------------- Container Ping ----------------
	widgetPingStatus := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	buttonPing := widget.NewButton("Ping Server", func() {
		log.Println("Start pinging server (IP: " + inputIP.Text + ")")
		recieved, err := ping(inputIP.Text)
		if err != nil || !recieved {
			widgetPingStatus.SetText("NOK")
		} else {
			widgetPingStatus.SetText("OK")
		}
	})

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
				setDefaultConf()
				//TODO find an better way for update the variables
				inputIP.SetText(strconv.Itoa(dstPort))
				inputPort.SetText(strconv.Itoa(dstPort))
				//TODO not possible to go back to default choice
				// selectProtocolTpye.SetSelected("Select one")
				widgetPingStatus.SetText("")
			}),
			fyne.NewMenuItem("Import configuration", func() {})),
		fyne.NewMenu("Preload Configuration",
			fyne.NewMenuItem("Warcraft 3", func() {
				setWar3Conf()
				//TODO find an better way for update the variables
				inputPort.SetText(strconv.Itoa(dstPort))
				selectProtocolTpye.SetSelected(protocoltpye)
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
