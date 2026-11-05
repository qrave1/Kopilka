package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func (s *screen) openingCash() {
	hint := widget.NewLabel("Введи наличные, которые были до первого обмена. Если были только рубли, заполни RUB. Уже сделанный обмен добавь отдельной записью с его датой.")
	hint.Wrapping = fyne.TextWrapWord
	fields := container.NewVBox(hint)
	opening := make(map[string]*widget.Entry, len(currencies))
	for _, currency := range currencies {
		field := entry(decimalText(s.ledger.Opening[currency], 2), "0,00")
		opening[currency] = field
		fields.Add(container.NewBorder(nil, nil, widget.NewLabel(currency), nil, field))
	}
	s.form("Начальные наличные", fields, s.home, func() {
		next, err := s.ledger.WithOpeningCash(inputValues(opening))
		if err != nil {
			dialog.ShowError(err, s.window)
			return
		}
		s.saveWithBalanceCheck(next)
	}, nil)
}

func (s *screen) settings() {
	fields := container.NewVBox()
	rates := make(map[string]*widget.Entry)
	original := make(map[string]string)
	beforeLeave := func(open func()) {
		for currency, field := range rates {
			if field.Text != original[currency] {
				dialog.ShowConfirm("Несохранённые курсы", "Выйти без сохранения изменений?", func(ok bool) {
					if ok {
						open()
					}
				}, s.window)
				return
			}
		}
		open()
	}
	fields.Add(widget.NewButton("Начальные наличные", s.afterTap(func() { beforeLeave(s.openingCash) })))
	for _, currency := range []string{"THB", "USD"} {
		value := ""
		if s.ledger.Rates[currency] > 0 {
			value = decimalText(s.ledger.Rates[currency], 6)
		}
		rates[currency] = entry(value, "₽ за единицу")
		original[currency] = value
		fields.Add(container.NewVBox(widget.NewLabel("Курс "+currency+" в рублях"), rates[currency]))
	}
	for _, control := range s.exportControls() {
		fields.Add(control)
	}

	save := widget.NewButton("Сохранить", func() {
		next, err := s.ledger.WithRates(inputValues(rates))
		if err != nil {
			dialog.ShowError(err, s.window)
			return
		}
		if s.persist(next) {
			s.settings()
		}
	})
	save.Importance = widget.HighImportance
	save.OnTapped = s.afterTap(save.OnTapped)
	s.showPageWithNavigation(settingsPage, widget.NewLabel("Настройки"), save, container.NewVScroll(fields), beforeLeave)
}

func inputValues(fields map[string]*widget.Entry) map[string]string {
	values := make(map[string]string, len(fields))
	for currency, field := range fields {
		values[currency] = field.Text
	}
	return values
}
