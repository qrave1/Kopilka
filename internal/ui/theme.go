package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type appTheme struct{ fyne.Theme }

func newAppTheme() fyne.Theme {
	return appTheme{theme.DefaultTheme()}
}

func (t appTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == categoryGlyphColor {
		return color.White
	}
	dark := variant != theme.VariantLight
	if dark {
		switch name {
		case theme.ColorNameBackground:
			return color.NRGBA{R: 0x12, G: 0x13, B: 0x17, A: 0xFF}
		case theme.ColorNameButton:
			return color.NRGBA{R: 0x22, G: 0x23, B: 0x2A, A: 0xFF}
		case theme.ColorNameInputBackground:
			return color.NRGBA{R: 0x18, G: 0x19, B: 0x20, A: 0xFF}
		}
	}
	if name == theme.ColorNamePressed {
		if dark {
			return color.NRGBA{R: 0xDC, G: 0xCB, B: 0xFF, A: 0xA0}
		}
		return color.NRGBA{R: 0x86, G: 0x61, B: 0xC9, A: 0x90}
	}
	if name == theme.ColorNamePrimary {
		if dark {
			return color.NRGBA{R: 0xB9, G: 0xA7, B: 0xE5, A: 0xFF}
		}
		return color.NRGBA{R: 0x5B, G: 0x46, B: 0x83, A: 0xFF}
	}
	if name == theme.ColorNameForegroundOnPrimary {
		if dark {
			return color.NRGBA{R: 0x20, G: 0x17, B: 0x31, A: 0xFF}
		}
		return color.White
	}
	return t.Theme.Color(name, variant)
}
