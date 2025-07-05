package gui

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"github.com/KingKeule/VPNubt/pkg/service"
)

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
				widget.NewFormItem("CIDR", &widget.Entry{Text: "10.0.0.0/24"}),
			),
		),
		widget.NewGroup("This Peer",
			widget.NewForm(
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
			),
		),
		widget.NewGroup("Other Peers",
			fyne.NewContainerWithLayout(
				layout.NewGridLayout(2),
				widget.NewButton("Add New", func() {}),
				widget.NewButton("Remove Selected", func() {}),
			),
			fyne.NewContainerWithLayout(
				layout.NewGridLayout(2),
				&widget.Entry{PlaceHolder: "Public IP"},
				&widget.Entry{PlaceHolder: "Public Key"},
			),
		),
	)
}
