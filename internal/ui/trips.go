package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"kopilka/internal/service"
)

func (s *screen) trips() {
	add := widget.NewButtonWithIcon("Новая поездка", theme.ContentAddIcon(), s.afterTap(s.newTrip))
	add.Importance = widget.HighImportance
	active, archived := container.NewVBox(), container.NewVBox()
	for _, journal := range s.ledger.Journals {
		if journal.Kind != service.TripJournal {
			continue
		}
		open := widget.NewButtonWithIcon(journal.Name, suitcaseIcon, s.afterTap(func() {
			s.journalID = journal.ID
			s.home()
		}))
		status := "Завершить"
		statusIcon := theme.ConfirmIcon()
		if journal.Archived {
			status = "Возобновить"
			statusIcon = theme.ViewRefreshIcon()
		}
		statusButton := widget.NewButtonWithIcon(status, statusIcon, func() { s.changeTripStatus(journal) })
		statusButton.Importance = widget.LowImportance
		row := surface(container.NewVBox(
			open,
			widget.NewLabel("Потрачено: "+money(s.ledger.TotalSpentRubles(journal.ID), "RUB")),
			statusButton,
		))
		if journal.Archived {
			archived.Add(row)
		} else {
			active.Add(row)
		}
	}
	if len(active.Objects) == 0 {
		message := widget.NewLabel("Создай поездку, чтобы вести её расходы и наличные отдельно.")
		message.Wrapping = fyne.TextWrapWord
		active.Add(message)
	}
	if len(archived.Objects) > 0 {
		active.Add(widget.NewSeparator())
		active.Add(widget.NewLabel("Завершённые поездки"))
		active.Add(archived)
	}
	s.settingsPage("Поездки", container.NewVBox(add, active))
}

func (s *screen) newTrip() {
	name := entry("", "Например, Таиланд")
	currency := widget.NewSelect(currencies, nil)
	currency.SetSelected("THB")
	fields := container.NewVBox(widget.NewLabel("Название"), name,
		widget.NewLabel("Основная валюта"), currency)
	s.form("Новая поездка", fields, s.trips, func() {
		next, err := s.ledger.WithTrip(name.Text, currency.Selected)
		if err != nil {
			dialog.ShowError(err, s.window)
			return
		}
		if s.persist(next) {
			s.journalID = next.Journals[len(next.Journals)-1].ID
			s.home()
		}
	}, nil)
}

func (s *screen) changeTripStatus(journal service.Journal) {
	save := func() {
		journal.Archived = !journal.Archived
		if s.persist(s.ledger.WithJournal(journal)) {
			if journal.Archived && s.journalID == journal.ID {
				s.journalID = 0
			}
			s.trips()
		}
	}
	if journal.Archived {
		save()
		return
	}
	dialog.ShowConfirm("Завершить поездку?", "Журнал и остатки останутся в завершённых поездках.", func(ok bool) {
		if ok {
			save()
		}
	}, s.window)
}
