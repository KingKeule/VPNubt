package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/KingKeule/VPNubt/pkg/config"
	"github.com/KingKeule/VPNubt/pkg/service"
)

var otherPeersSection *fyne.Container

func wireguard() *fyne.Container {
	return container.New(
		layout.NewVBoxLayout(),
		widget.NewCard("Control", "",
			container.New(
				layout.NewGridLayout(2),
				widget.NewButton("Start / Stop", func() {}),
				&canvas.Rectangle{
					FillColor:   color.RGBA{R: 0, G: 0, B: 0, A: 50.0},
					StrokeColor: color.Black,
				},
			),
		),
		widget.NewCard("IPv4 Network", "",
			widget.NewForm(
				widget.NewFormItem("CIDR", &widget.Entry{Text: config.Prefs.String("network")}),
			),
		),
		widget.NewCard("This Peer", "",
			widget.NewForm(
				widget.NewFormItem("Local IP", &widget.Select{
					Selected: config.Prefs.String("thisPeerIP"),
					Options:  config.ThisNetworkHostAddresses(),
					OnChanged: func(s string) {
						config.Prefs.SetString("thisPeerIP", s)
						// TODO warn if overwrite existing peer
						updateOtherPeerSection()
					},
				}),
				widget.NewFormItem("Public IP", &widget.Entry{
					PlaceHolder: "TODO",
				}),
				widget.NewFormItem("Private Key", &widget.Entry{
					Text: service.WgConf().Interface.PrivateKey.String(),
				}),
				widget.NewFormItem("Public Key", &widget.Entry{
					Text: service.WgConf().Interface.PrivateKey.Public().String(),
				}),
				// TODO share button
			),
		),
		widget.NewCard("Other Peers", "",
			widget.NewForm(
				widget.NewFormItem("Local IP", &widget.Select{
					Options:   config.ThisNetworkHostAddresses(),
					OnChanged: func(s string) {},
				}),
				widget.NewFormItem("Public IP", &widget.Entry{
					PlaceHolder: "Public IP",
				}),
				widget.NewFormItem("Public Key", &widget.Entry{
					PlaceHolder: "Public Key",
				}),
			),
		),
		//widget.NewCard("Other Peers", "", createOtherPeersSection()),
	)
}

func createOtherPeersSection() *fyne.Container {

	objects := make([]fyne.CanvasObject, 0)

	for _, h := range config.ThisNetworkHostAddresses() {

		la := container.NewGridWithRows(3)
		laFirst := container.NewGridWithColumns(2)

		entry := widget.NewEntry()
		entry.SetPlaceHolder("Local IP")
		entry.SetText(h)
		laFirst.Add(entry)

		button := widget.NewButton("Apply", func() {})
		laFirst.Add(button)

		la.Add(laFirst)

		entry = widget.NewEntry()
		entry.SetPlaceHolder("Public IP")
		la.Add(entry)

		entry = widget.NewEntry()
		entry.SetPlaceHolder("Public Key")
		la.Add(entry)

		objects = append(objects, la)
		objects = append(objects, widget.NewSeparator())
	}

	otherPeersSection = container.NewVBox(objects...)

	//updateOtherPeerSection()

	return otherPeersSection
}

func updateOtherPeerSection() {
	for _, o := range otherPeersSection.Objects {
		toDisable := o.(*fyne.Container).Objects
		privateIP := toDisable[0].(*widget.Entry)
		publicIP := toDisable[1].(*widget.Entry)
		publicKey := toDisable[2].(*widget.Entry)
		apply := toDisable[3].(*widget.Button)
		if privateIP.Text == config.Prefs.String("thisPeerIP") {
			privateIP.Disable()
			publicIP.Disable()
			publicKey.Disable()
			apply.Disable()
			continue
		}
		privateIP.Enable()
		publicIP.Enable()
		publicKey.Enable()
		apply.Enable()
	}
}
