package ui

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func surface(content fyne.CanvasObject) *fyne.Container {
	background := canvas.NewRectangle(theme.Color(theme.ColorNameButton))
	background.CornerRadius = 16
	return container.NewStack(background, container.NewPadded(content))
}

func sectionTitle(text string) *widget.RichText {
	title := widget.NewRichText(&widget.TextSegment{Text: text, Style: widget.RichTextStyleHeading})
	title.Truncation = fyne.TextTruncateEllipsis
	return title
}

func iconBadge(icon fyne.Resource) *fyne.Container {
	glyph := canvas.NewImageFromResource(theme.NewPrimaryThemedResource(icon))
	glyph.FillMode = canvas.ImageFillContain
	glyph.SetMinSize(fyne.NewSize(22, 22))
	return container.NewPadded(container.NewCenter(glyph))
}

type iconTile struct {
	widget.BaseWidget
	label    string
	icon     fyne.Resource
	active   bool
	onTapped func()
	press    *canvas.Rectangle
	pulse    *fyne.Animation
}

func newIconTile(label string, icon fyne.Resource, active bool, tapped func()) *iconTile {
	tile := &iconTile{label: label, icon: icon, active: active, onTapped: tapped}
	tile.ExtendBaseWidget(tile)
	return tile
}

func (t *iconTile) CreateRenderer() fyne.WidgetRenderer {
	ink := theme.ColorNameForeground
	if t.active {
		ink = theme.ColorNamePrimary
	}
	glyph := canvas.NewImageFromResource(theme.NewColoredResource(t.icon, ink))
	glyph.FillMode = canvas.ImageFillContain
	glyph.SetMinSize(fyne.NewSize(22, 22))
	caption := canvas.NewText(t.label, theme.Color(ink))
	caption.Alignment = fyne.TextAlignCenter
	caption.TextSize = 12
	caption.TextStyle.Bold = t.active
	t.press = canvas.NewRectangle(color.Transparent)
	t.press.CornerRadius = 12
	content := container.NewVBox(container.NewCenter(glyph), caption)
	return widget.NewSimpleRenderer(container.NewStack(t.press, container.NewPadded(content)))
}

func (t *iconTile) Tapped(*fyne.PointEvent) {
	if t.onTapped == nil {
		return
	}
	if t.press != nil && fyne.CurrentApp().Settings().ShowAnimations() {
		if t.pulse != nil {
			t.pulse.Stop()
		}
		ink := color.NRGBAModel.Convert(theme.Color(theme.ColorNamePressed)).(color.NRGBA)
		t.pulse = fyne.NewAnimation(180*time.Millisecond, func(progress float32) {
			t.press.FillColor = color.NRGBA{R: ink.R, G: ink.G, B: ink.B, A: uint8(float32(ink.A) * (1 - progress))}
			t.press.Refresh()
		})
		t.pulse.Start()
	}
	t.onTapped()
}

func vectorIcon(name, path string) fyne.Resource {
	svg := fyne.NewStaticResource(name+".svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="#000" fill-rule="evenodd" d="`+path+`"/></svg>`))
	return theme.NewThemedResource(svg)
}

var (
	suitcaseIcon = vectorIcon("suitcase", "M3 7h5V4h8v3h5v14H3z M10 6v1h4V6z M7 9v10h2V9z M15 9v10h2V9z")
	walletIcon   = vectorIcon("wallet", "M3 5h17v4h2v8h-2v3H3z M14 11v4h6v-4z")
	swapIcon     = vectorIcon("swap", "M3 7h12V3l6 5-6 5V9H3z M21 17H9v4l-6-5 6-5v4h12z")
	tagIcon      = vectorIcon("tag", "M3 3h8l10 10-8 8L3 11z M6 6v3h3V6z")
)

func categoryIcon(name string) fyne.Resource {
	switch name {
	case "Еда":
		return foodIcon
	case "Транспорт":
		return transportIcon
	case "Жильё":
		return theme.HomeIcon()
	case "Развлечения":
		return entertainmentIcon
	case "Покупки":
		return shoppingIcon
	default:
		return tagIcon
	}
}

var (
	foodIcon          = vectorIcon("food", "M3 3h2v5h1V3h2v5h1V3h2v6a4 4 0 0 1-3 4v8H6v-8a4 4 0 0 1-3-4z M19 3v18h-2v-7h-4v-3a8 8 0 0 1 6-8z")
	transportIcon     = vectorIcon("transport", "M3 8a5 5 0 0 1 5-5h8a5 5 0 0 1 5 5v10h-2v3h-3v-3H8v3H5v-3H3z M5 8v3h14V8z M6 14v2h3v-2z M15 14v2h3v-2z")
	entertainmentIcon = vectorIcon("entertainment", "M12 3l2.5 5 5.5.8-4 3.9.9 5.6L12 15.7l-4.9 2.6.9-5.6-4-3.9 5.5-.8z")
	shoppingIcon      = vectorIcon("shopping", "M4 8h4V6a4 4 0 0 1 8 0v2h4l1 13H3z M10 8h4V6a2 2 0 0 0-4 0z")
)
