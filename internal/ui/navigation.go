package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type page int

const (
	journalPage page = iota
	tripsPage
	categoriesPage
	settingsPage
)

func (s *screen) showPage(active page, header, footer, body fyne.CanvasObject) {
	s.showPageWithNavigation(active, header, footer, body, nil)
}

func (s *screen) showPageWithNavigation(active page, header, actions, body fyne.CanvasObject, beforeLeave func(func())) {
	tabs := []struct {
		label string
		icon  fyne.Resource
		open  func()
	}{
		{"Журнал", theme.HomeIcon(), s.home},
		{"Поездки", suitcaseIcon, s.trips},
		{"Категории", tagIcon, s.categories},
		{"Настройки", theme.SettingsIcon(), s.settings},
	}
	bar := container.NewGridWithColumns(len(tabs))
	for i, tab := range tabs {
		open := tab.open
		if beforeLeave != nil {
			open = func() { beforeLeave(tab.open) }
		}
		var tapped func()
		if page(i) == active {
			// Tapping the current tab must not reset fields or scroll position.
		} else {
			tapped = s.afterTap(open)
		}
		bar.Add(newIconTile(tab.label, tab.icon, page(i) == active, tapped))
	}
	footer := container.NewVBox()
	if actions != nil {
		footer.Add(actions)
	}
	footer.Add(widget.NewSeparator())
	footer.Add(surface(bar))
	s.window.SetContent(container.NewBorder(header, footer, nil, nil, body))
}
