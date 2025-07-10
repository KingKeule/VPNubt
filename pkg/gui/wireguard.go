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

var otherPeersIP *widget.Select = nil

func wireguard() *fyne.Container {

	otherPeersIP = widget.NewSelect(config.ThisNetworkRemainingHostAddresses(), func(s string) {})

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
						otherPeersIP.Options = config.ThisNetworkRemainingHostAddresses()
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
				widget.NewFormItem("Local IP", otherPeersIP),
				widget.NewFormItem("Public IP", &widget.Entry{
					PlaceHolder: "Public IP",
				}),
				widget.NewFormItem("Public Key", &widget.Entry{
					PlaceHolder: "Public Key",
				}),
			),
		),
	)
}
