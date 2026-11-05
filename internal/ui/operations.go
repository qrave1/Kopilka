package ui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"kopilka/internal/service"
)

func (s *screen) editOperation(index int, exchange bool) {
	var previous service.Operation
	if index >= 0 {
		previous = s.ledger.Operations[index]
	}
	date := entry(time.Now().Format(time.DateOnly), "ГГГГ-ММ-ДД")
	amount := entry("", "0,00")
	currency := widget.NewSelect(currencies, nil)
	currency.SetSelected("THB")
	received := widget.NewSelect(currencies, nil)
	received.SetSelected("USD")
	if exchange {
		from, to := s.ledger.SuggestedExchange()
		currency.SetSelected(from)
		received.SetSelected(to)
	}
	if index >= 0 {
		date.SetText(previous.Date)
		amount.SetText(decimalText(previous.Amount, 2))
		currency.SetSelected(previous.Currency)
	}
	fields := container.NewVBox(widget.NewLabel("Дата"), date)
	preview := widget.NewLabel("")
	preview.Wrapping = fyne.TextWrapWord
	names, categoryIDs := s.ledger.ExpenseCategories(previous.CategoryID)
	category := widget.NewSelect(names, nil)
	category.PlaceHolder = "Выбери категорию"
	incoming := entry("", "0,00")
	title := "Трата"
	if exchange {
		title = "Обмен"
		if index >= 0 {
			received.SetSelected(previous.Received)
			incoming.SetText(decimalText(previous.Incoming, 2))
		}
		fields.Add(container.NewVBox(widget.NewLabel("Отдал"), currency, amount,
			widget.NewLabel("Получил"), received, incoming))
	} else {
		category.SetSelected(s.ledger.CategoryName(previous.CategoryID))
		fields.Add(container.NewVBox(widget.NewLabel("Сумма"), currency, amount,
			widget.NewLabel("Категория"), category, preview))
		if len(names) == 0 {
			fields.Add(widget.NewButton("Добавить категорию", s.categories))
		}
		updatePreview := func() {
			value, err := service.ParseDecimal(amount.Text, 2)
			if err != nil || value == 0 {
				preview.SetText("")
				return
			}
			op, err := s.ledger.Expense(date.Text, currency.Selected, categoryIDs[category.Selected], value, previous)
			if err != nil {
				preview.SetText(err.Error())
				return
			}
			preview.SetText("~ " + money(op.Rubles, "RUB"))
		}
		amount.OnChanged = func(string) { updatePreview() }
		currency.OnChanged = func(string) { updatePreview() }
		updatePreview()
	}

	save := func() {
		op, err := s.ledger.PrepareOperation(service.OperationInput{
			Date: date.Text, Currency: currency.Selected, Amount: amount.Text,
			Received: received.Selected, Incoming: incoming.Text,
			CategoryID: categoryIDs[category.Selected], Exchange: exchange,
		}, previous)
		if err != nil {
			dialog.ShowError(err, s.window)
			return
		}
		s.saveWithBalanceCheck(s.ledger.WithOperation(index, op))
	}
	var remove func()
	if index >= 0 {
		remove = func() {
			dialog.ShowConfirm("Удалить запись?", "Остатки будут пересчитаны.", func(ok bool) {
				if ok {
					s.saveWithBalanceCheck(s.ledger.WithoutOperation(index))
				}
			}, s.window)
		}
	}
	s.form(title, fields, s.home, save, remove)
}
