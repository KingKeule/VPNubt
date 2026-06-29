package gui

// Info: https://developer.fyne.io/extend/custom-theme
// Thanks to https://github.com/lusingander/fyne-theme-generator

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

var _ fyne.Theme = (*ThemeCustom)(nil)

type ThemeCustom struct {
	ThemeType string
}

func (cTheme ThemeCustom) Color(col fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch cTheme.ThemeType {
	case "Grey":
		variant = theme.VariantDark
		switch col {
		case theme.ColorNameBackground: // general background
			return color.NRGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff}
		case theme.ColorNameForeground: // general foreground
			return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
		case theme.ColorNameInputBackground: // entry, checkbox, slider line background
			return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x30}
		case theme.ColorNameOverlayBackground: // dialog background
			return color.NRGBA{R: 0x36, G: 0x36, B: 0x36, A: 0xff}
		case theme.ColorNamePlaceHolder: // text placeholder
			return color.NRGBA{R: 0xb2, G: 0xb2, B: 0xb2, A: 0xff}
		case theme.ColorNameMenuBackground: // menu background
			return color.NRGBA{R: 0x36, G: 0x36, B: 0x36, A: 0xff}
		case theme.ColorNameButton: // button background
			return color.NRGBA{R: 0x50, G: 0x50, B: 0x50, A: 0x70}
		case theme.ColorNameHover: // hover effect
			return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x20}
		case theme.ColorNameScrollBar: //scrollbar
			return color.NRGBA{R: 0x70, G: 0x70, B: 0x70, A: 0x99}
		case theme.ColorNamePrimary: // progress bar etc.
			return color.NRGBA{R: 0x21, G: 0x96, B: 0xf3, A: 0xff}
		case theme.ColorNameSeparator: // seperator, table cell lines
			return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x6f}
		case theme.ColorNameInputBorder: // checkbox & input border
			return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x30}
		case theme.ColorNameError: // checkbox & input border
			return color.NRGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}
		}
	case "Bright":
		variant = theme.VariantLight
		switch col {
		case theme.ColorNameForeground: // general foreground
			return color.NRGBA{R: 0x20, G: 0x20, B: 0x20, A: 0xff}
		case theme.ColorNameInputBackground: // entry, checkbox background
			return color.NRGBA{R: 0x80, G: 0x80, B: 0x80, A: 0x60}
		case theme.ColorNameButton: // button background
			return color.NRGBA{R: 0x44, G: 0x44, B: 0x44, A: 0x00}
		case theme.ColorNameSeparator: // seperator, table cell lines
			return color.NRGBA{R: 0x60, G: 0x60, B: 0x60, A: 0xff}
		}
	}

	return theme.DefaultTheme().Color(col, variant)
}

func (m ThemeCustom) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m ThemeCustom) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m ThemeCustom) Size(s fyne.ThemeSizeName) float32 {
	switch s {
	case theme.SizeNamePadding:
		return 4
	case theme.SizeNameText:
		return 14
	case theme.SizeNameScrollBar:
		return 12
	case theme.SizeNameScrollBarSmall:
		return 4
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameInputBorder:
		return 2
	case theme.SizeNameInputRadius:
		return 3
	case theme.SizeNameSelectionRadius:
		return 0
	default:
		return theme.DefaultTheme().Size(s)
	}
}
