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
	if name == theme.ColorNamePressed {
		if variant == theme.VariantDark {
			return color.NRGBA{R: 0xDC, G: 0xCB, B: 0xFF, A: 0xA0}
		}
		return color.NRGBA{R: 0x86, G: 0x61, B: 0xC9, A: 0x90}
	}
	if name == theme.ColorNamePrimary {
		if variant == theme.VariantDark {
			return color.NRGBA{R: 0xB9, G: 0xA7, B: 0xE5, A: 0xFF}
		}
		return color.NRGBA{R: 0x5B, G: 0x46, B: 0x83, A: 0xFF}
	}
	if name == theme.ColorNameForegroundOnPrimary {
		if variant == theme.VariantDark {
			return color.NRGBA{R: 0x20, G: 0x17, B: 0x31, A: 0xFF}
		}
		return color.White
	}
	return t.Theme.Color(name, variant)
}
