package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/KingKeule/VPNubt/pkg/config"
	"github.com/KingKeule/VPNubt/pkg/service"
)

var currentOtherPeerLocalIP binding.String
var currentOtherPeerPublicIP string
var currentOtherPeerPublicKey string

func wireguard() *fyne.Container {

	selectedOtherPeer := config.FirstOfRemainingHostAddresses()

	currentOtherPeerLocalIP = binding.NewString()
	currentOtherPeerLocalIP.Set(selectedOtherPeer)

	currentOtherPeerPublicIP = config.GetPublicIPOfPeer(selectedOtherPeer)
	currentOtherPeerPublicKey = config.GetPublicKeyOfPeer(selectedOtherPeer)

	otherPeerLocalIp := &widget.Select{
		Options:  config.ThisNetworkRemainingHostAddresses(),
		Selected: selectedOtherPeer,
		OnChanged: func(updatedValue string) {
			currentOtherPeerLocalIP.Set(updatedValue)
		},
	}

	otherPeerPublicIP := &widget.Entry{
		PlaceHolder: "Public IP",
		Text:        config.GetPublicIPOfPeer(selectedOtherPeer),
		OnChanged: func(updatedValue string) {
			currentOtherPeerPublicIP = updatedValue
		},
	}

	otherPeerPublicKey := &widget.Entry{
		PlaceHolder: "Public Key",
		Text:        config.GetPublicKeyOfPeer(selectedOtherPeer),
		OnChanged: func(updatedValue string) {
			currentOtherPeerPublicKey = updatedValue
		},
	}

	currentOtherPeerLocalIP.AddListener(binding.NewDataListener(func() {
		peer, _ := currentOtherPeerLocalIP.Get()
		currentOtherPeerPublicIP = config.GetPublicIPOfPeer(peer)
		currentOtherPeerPublicKey = config.GetPublicKeyOfPeer(peer)
		otherPeerPublicIP.Text = currentOtherPeerPublicIP
		otherPeerPublicKey.Text = currentOtherPeerPublicKey
		otherPeerPublicIP.Refresh()
		otherPeerPublicKey.Refresh()
	}))

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
				widget.NewFormItem("CIDR", &widget.Entry{
					Text: config.Prefs.String("network"),
				}),
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
			container.New(
				layout.NewVBoxLayout(),
				widget.NewForm(
					widget.NewFormItem("Local IP", otherPeerLocalIp),
					widget.NewFormItem("Public IP", otherPeerPublicIP),
					widget.NewFormItem("Public Key", otherPeerPublicKey),
				),
				widget.NewButton("Apply Peer Settings", func() {
					peer, _ := currentOtherPeerLocalIP.Get()
					config.SetPublicIPOfPeer(peer, currentOtherPeerPublicIP)
					config.SetPublicKeyOfPeer(peer, currentOtherPeerPublicKey)
				}),
			),
		),
	)
}
