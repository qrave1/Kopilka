package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"kopilka/internal/service"
)

var currencies = service.Currencies()

type screen struct {
	window     fyne.Window
	save       func(service.Ledger) error
	ledger     service.Ledger
	tapPending bool
}

func Run(a fyne.App, l service.Ledger, save func(service.Ledger) error, loadErr error) {
	a.Settings().SetTheme(newAppTheme())
	w := a.NewWindow("Копилка")
	if loadErr != nil {
		w.SetContent(widget.NewLabel("Не удалось открыть журнал."))
		dialog.ShowError(loadErr, w)
	} else {
		s := &screen{window: w, save: save, ledger: l}
		s.home()
	}
	w.ShowAndRun()
}

func money(amount int64, currency string) string {
	symbols := map[string]string{"RUB": "₽", "USD": "$", "THB": "THB"}
	return moneyAmount(amount) + " " + symbols[currency]
}

func moneyAmount(amount int64) string {
	sign := ""
	if amount < 0 {
		sign = "-"
		amount = -amount
	}
	return fmt.Sprintf("%s%d,%02d", sign, amount/100, amount%100)
}

func (s *screen) persist(next service.Ledger) bool {
	if err := s.save(next); err != nil {
		dialog.ShowError(fmt.Errorf("не удалось сохранить журнал: %w", err), s.window)
		return false
	}
	s.ledger = next
	return true
}

func (s *screen) commit(next service.Ledger) {
	if s.persist(next) {
		s.home()
	}
}

func (s *screen) saveWithBalanceCheck(next service.Ledger) {
	var warnings []string
	for _, balance := range s.ledger.WorsenedNegativeBalances(next) {
		warnings = append(warnings, balance.Currency+" "+money(balance.Amount, balance.Currency))
	}
	if len(warnings) > 0 {
		dialog.ShowConfirm("Отрицательный остаток", strings.Join(warnings, "\n")+"\nСохранить изменение?", func(ok bool) {
			if ok {
				s.commit(next)
			}
		}, s.window)
		return
	}
	s.commit(next)
}

func entry(text, hint string) *widget.Entry {
	field := widget.NewEntry()
	field.SetPlaceHolder(hint)
	field.SetText(text)
	return field
}

func decimalText(amount int64, places int) string {
	scale := int64(100)
	if places == 6 {
		scale = service.RateScale
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%d,%0*d", amount/scale, places, amount%scale), "0"), ",")
}

func (s *screen) form(title string, fields *fyne.Container, back, save, remove func()) {
	button := widget.NewButton("Сохранить", s.afterTap(save))
	button.Importance = widget.HighImportance
	actions := container.NewVBox(button)
	if remove != nil {
		actions.Add(widget.NewButton("Удалить запись", remove))
	}
	top := container.NewVBox(widget.NewButton("Назад", s.afterTap(back)), widget.NewLabel(title))
	s.window.SetContent(container.NewBorder(top, actions, nil, nil, container.NewVScroll(fields)))
}
