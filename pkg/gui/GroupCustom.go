package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type GroupCustom struct {
	widget.BaseWidget
	HeadlineText string
}

func NewGroupCustom(headlineText string) *GroupCustom {
	groupCustom := &GroupCustom{
		HeadlineText: headlineText,
	}
	groupCustom.ExtendBaseWidget(groupCustom)

	return groupCustom
}

type customGroupRenderer struct {
	groupCustom        *GroupCustom
	frameBG, textFrame *canvas.Rectangle
	separatorLine      *canvas.Line
	text               *canvas.Text
}

func (groupCustom *GroupCustom) CreateRenderer() fyne.WidgetRenderer {
	groupCustomColorBG := theme.BackgroundColor()

	return &customGroupRenderer{
		groupCustom:   groupCustom,
		frameBG:       canvas.NewRectangle(groupCustomColorBG),
		textFrame:     canvas.NewRectangle(groupCustomColorBG),
		separatorLine: canvas.NewLine(color.NRGBA{R: 30, G: 30, B: 30, A: 255}),
		text:          canvas.NewText(groupCustom.HeadlineText, theme.ForegroundColor()),
	}
}

func (r *customGroupRenderer) Layout(s fyne.Size) {
	// Measure the size of the text so we can calculate the center offset.
	ts := fyne.MeasureText(r.text.Text, r.text.TextSize, r.text.TextStyle)

	// Center the headline text
	r.text.Move(fyne.Position{X: (s.Width - ts.Width) / 2, Y: (s.Height - ts.Height) / 2})

	//Center the separator line of the head text
	r.separatorLine.StrokeWidth = 3
	r.separatorLine.Move(fyne.NewPos(0, ((s.Height-ts.Height)/2)+11))
	r.separatorLine.Resize(fyne.NewSize(s.Width, 0))

	// Adjust the frame for the text otherwise the separator line would cross out the text
	var textFrameSpace float32 = 4
	r.textFrame.Move(fyne.Position{X: ((s.Width - ts.Width) / 2) - textFrameSpace, Y: ((s.Height - ts.Height) / 2)})
	r.textFrame.Resize(fyne.NewSize(ts.Width+2*textFrameSpace, ts.Height))

	// Expand the background frame
	r.frameBG.Resize(s)
}

func (r *customGroupRenderer) Refresh() {
	r.frameBG.Refresh()
	r.separatorLine.Refresh()
	r.textFrame.Refresh()
	r.text.Refresh()
}

func (r *customGroupRenderer) MinSize() fyne.Size {
	// Measure the size of the text so we can calculate a border size.
	ts := fyne.MeasureText(r.text.Text, r.text.TextSize, r.text.TextStyle)

	// Use the theme padding to set a border size
	return fyne.NewSize(ts.Width+theme.Padding()*2, ts.Height+theme.Padding()*2)
}

// Define the order in which the objects are placed on top of each other.
// The first object is the lowest and the last the highest.
func (r *customGroupRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.frameBG, r.separatorLine, r.textFrame, r.text}
}

func (r *customGroupRenderer) Destroy() {}
