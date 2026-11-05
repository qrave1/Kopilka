package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type categoryIconOption struct {
	key   string
	label string
	icon  fyne.Resource
	rgb   uint32
}

var categoryIcons = []categoryIconOption{
	{"tag", "Другое", tagIcon, 0x765044},
	{"bills", "Счета", vectorIcon("bills", "M13 2L4 14h7l-1 8 10-12h-7z"), 0x117B9F},
	{"clothes", "Одежда", vectorIcon("shirt", "M7 3h3a2 2 0 0 0 4 0h3l5 4-3 5-2-1v10H7V11l-2 1-3-5z"), 0x0869A7},
	{"education", "Образование", vectorIcon("school", "M12 3L1 9l11 6 11-6z M5 13l7 4 7-4v5l-7 4-7-4z"), 0x067B77},
	{"star", "Развлечения", entertainmentIcon, 0x007E83},
	{"fitness", "Фитнес", vectorIcon("fitness", "M2 9h2v6H2z M5 6h3v12H5z M9 11h6v2H9z M16 6h3v12h-3z M20 9h2v6h-2z"), 0x507A0D},
	{"food", "Продукты", foodIcon, 0x876100},
	{"gift", "Подарки", vectorIcon("gift", "M3 8h18v5H3z M4 14h7v7H4z M13 14h7v7h-7z M12 8C3 9 5 1 9 3l3 4 3-4c4-2 6 6-3 5z"), 0x9B5A00},
	{"health", "Здоровье", vectorIcon("health", "M9 3h6v6h6v6h-6v6H9v-6H3V9h6z"), 0xBD3535},
	{"furniture", "Мебель", vectorIcon("sofa", "M5 5h14v6h-3v4H8v-4H5z M2 12h4v5h12v-5h4v8h-2v2h-2v-2H6v2H4v-2H2z"), 0xA31D67},
	{"pet", "Питомцы", vectorIcon("paw", "M4 6a2 3 0 1 0 0 6 2 3 0 1 0 0-6z M9 3a2 3 0 1 0 0 6 2 3 0 1 0 0-6z M15 3a2 3 0 1 0 0 6 2 3 0 1 0 0-6z M20 6a2 3 0 1 0 0 6 2 3 0 1 0 0-6z M12 11c3 0 7 4 7 7 0 4-5 2-7 2s-7 2-7-2c0-3 4-7 7-7z"), 0x68429E},
	{"shopping", "Покупки", shoppingIcon, 0x7740BC},
	{"transport", "Поездки", transportIcon, 0xAD196F},
	{"trip", "Путешествие", suitcaseIcon, 0x974509},
	{"home", "Жильё", theme.HomeIcon(), 0x395989},
	{"document", "Документы", theme.FileTextIcon(), 0x475569},
}

func categoryIcon(key string) categoryIconOption {
	for _, option := range categoryIcons {
		if option.key == key {
			return option
		}
	}
	return categoryIcons[0]
}

const categoryGlyphColor fyne.ThemeColorName = "kopilkaCategoryGlyph"

type categoryMark struct {
	circle *canvas.Circle
	glyph  *canvas.Image
}

func newCategoryMark(key string) categoryMark {
	mark := categoryMark{circle: canvas.NewCircle(color.Transparent), glyph: &canvas.Image{}}
	mark.glyph.FillMode = canvas.ImageFillContain
	mark.glyph.SetMinSize(fyne.NewSize(20, 20))
	mark.set(categoryIcon(key))
	return mark
}

func (m categoryMark) set(option categoryIconOption) {
	m.circle.FillColor = color.NRGBA{R: uint8(option.rgb >> 16), G: uint8(option.rgb >> 8), B: uint8(option.rgb), A: 255}
	m.glyph.Resource = theme.NewColoredResource(option.icon, categoryGlyphColor)
	m.circle.Refresh()
	m.glyph.Refresh()
}

func (m categoryMark) content() fyne.CanvasObject {
	return container.NewCenter(container.NewStack(m.circle, container.NewPadded(container.NewCenter(m.glyph))))
}

func categoryIconPicker(selected *string) *fyne.Container {
	previous := *selected
	*selected = "tag"
	for _, option := range categoryIcons {
		if option.key == previous {
			*selected = previous
			break
		}
	}
	grid := container.NewGridWithColumns(3)
	var refresh func()
	refresh = func() {
		grid.RemoveAll()
		for _, option := range categoryIcons {
			grid.Add(newIconTile(option.label, option.icon, option.key == *selected, func() {
				*selected = option.key
				refresh()
			}))
		}
	}
	refresh()
	return container.NewVBox(widget.NewLabel("Иконка категории"), surface(grid))
}
