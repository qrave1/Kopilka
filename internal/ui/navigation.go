package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type page int

const (
	journalPage page = iota
	categoriesPage
	settingsPage
)

func (s *screen) showPage(active page, header, footer, body fyne.CanvasObject) {
	s.showPageWithNavigation(active, header, footer, body, nil)
}

func (s *screen) showPageWithNavigation(active page, header, actions, body fyne.CanvasObject, beforeLeave func(func())) {
	tabs := []struct {
		label string
		open  func()
	}{
		{"Журнал", s.home},
		{"Категории", s.categories},
		{"Настройки", s.settings},
	}
	bar := container.NewGridWithColumns(len(tabs))
	for i, tab := range tabs {
		open := tab.open
		if beforeLeave != nil {
			open = func() { beforeLeave(tab.open) }
		}
		button := widget.NewButton(tab.label, s.afterTap(open))
		button.Importance = widget.LowImportance
		if page(i) == active {
			button.Importance = widget.HighImportance
			// Tapping the current tab must not reset fields or scroll position.
			button.OnTapped = nil
		}
		bar.Add(button)
	}
	footer := container.NewVBox()
	if actions != nil {
		footer.Add(actions)
	}
	footer.Add(widget.NewSeparator())
	footer.Add(bar)
	s.window.SetContent(container.NewBorder(header, footer, nil, nil, body))
}
