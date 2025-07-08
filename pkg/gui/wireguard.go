package gui

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/container"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"github.com/KingKeule/VPNubt/pkg/config"
	"github.com/KingKeule/VPNubt/pkg/service"
)

var otherPeersSection *fyne.Container

func wireguard() *fyne.Container {
	return fyne.NewContainerWithLayout(
		layout.NewVBoxLayout(),
		widget.NewGroup("Control",
			fyne.NewContainerWithLayout(
				layout.NewGridLayout(2),
				widget.NewButton("Start / Stop", func() {}),
				&canvas.Rectangle{
					FillColor:   color.RGBA{R: 0, G: 0, B: 0, A: 50.0},
					StrokeColor: color.Black,
				},
			),
		),
		widget.NewGroup("IPv4 Network",
			widget.NewForm(
				widget.NewFormItem("CIDR", &widget.Entry{Text: config.Prefs.String("network")}),
			),
		),
		widget.NewGroup("This Peer",
			widget.NewForm(
				widget.NewFormItem("IP", &widget.Select{
					Selected: config.Prefs.String("thisPeerIP"),
					Options:  config.ThisNetworkHostAddresses(),
					OnChanged: func(s string) {
						config.Prefs.SetString("thisPeerIP", s)
						// TODO warn if overwrite existing peer
						updateOtherPeerSection()
					},
				}),
				widget.NewFormItem("Private Key", &widget.Entry{
					Text:     service.WgConf().Interface.PrivateKey.String(),
					ReadOnly: true,
					Wrapping: fyne.TextTruncate,
				}),
				widget.NewFormItem("Public Key", &widget.Entry{
					Text:     service.WgConf().Interface.PrivateKey.Public().String(),
					ReadOnly: true,
					Wrapping: fyne.TextTruncate,
				}),
				// TODO share button
			),
		),
		widget.NewGroup("Other Peers", createOtherPeersSection()),
	)
}

func createOtherPeersSection() *fyne.Container {

	objects := make([]fyne.CanvasObject, 0)

	for _, h := range config.ThisNetworkHostAddresses() {

		la := fyne.NewContainerWithLayout(
			layout.NewGridLayout(4),
		)

		entry := widget.NewEntry()
		entry.SetPlaceHolder("Local IP")
		entry.SetText(h)
		// Wrap entry in a container that prevents resizing
		fixedWidth := fyne.NewSize(20, entry.MinSize().Height)
		wrapper := container.NewMax(entry)
		wrapper.Resize(fixedWidth)
		la.Add(wrapper)

		entry = widget.NewEntry()
		entry.SetPlaceHolder("Public IP")
		// Wrap entry in a container that prevents resizing
		fixedWidth = fyne.NewSize(20, entry.MinSize().Height)
		wrapper = container.NewMax(entry)
		wrapper.Resize(fixedWidth)
		la.Add(wrapper)

		entry = widget.NewEntry()
		entry.SetPlaceHolder("Public Key")
		// Wrap entry in a container that prevents resizing
		fixedWidth = fyne.NewSize(20, entry.MinSize().Height)
		wrapper = container.NewMax(entry)
		wrapper.Resize(fixedWidth)
		la.Add(wrapper)

		button := widget.NewButton("Apply", func() {})
		// Wrap entry in a container that prevents resizing
		fixedWidth = fyne.NewSize(20, entry.MinSize().Height)
		wrapper = container.NewMax(button)
		wrapper.Resize(fixedWidth)
		la.Add(wrapper)

		objects = append(objects, la)

		// &fyne.Container{
		// 	Layout: layout.NewGridLayout(4),
		// 	Objects: []fyne.CanvasObject{
		// 		&widget.Entry{
		// 			Text:        strings.ReplaceAll(h, ".", "_"),
		// 			PlaceHolder: "Custom Name",
		// 			Wrapping:    fyne.TextTruncate,
		// 		},
		// 		&widget.Entry{
		// 			PlaceHolder: "Public IP",
		// 			Wrapping:    fyne.TextTruncate,
		// 		},
		// 		&widget.Entry{
		// 			PlaceHolder: "Public Key",
		// 			Wrapping:    fyne.TextTruncate,
		// 			OnChanged:   func(s string) {},
		// 		},
		// 		&widget.Button{
		// 			Text:     "Apply",
		// 			OnTapped: func() {},
		// 		},
		// 	},
		// },

	}

	otherPeersSection = &fyne.Container{
		Layout:  layout.NewVBoxLayout(),
		Objects: objects,
	}

	updateOtherPeerSection()

	return otherPeersSection
}

func updateOtherPeerSection() {
	for _, o := range otherPeersSection.Objects {
		toDisable := o.(*fyne.Container).Objects
		privateIP := toDisable[0].(*fyne.Container).Objects[0].(*widget.Entry)
		publicIP := toDisable[1].(*fyne.Container).Objects[0].(*widget.Entry)
		publicKey := toDisable[2].(*fyne.Container).Objects[0].(*widget.Entry)
		apply := toDisable[3].(*fyne.Container).Objects[0].(*widget.Button)
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
