package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"kopilka/internal/service"
)

func (s *screen) converter() {
	amount := entry("", "0,00")
	manualRate := entry("", "По курсам из настроек")
	rateLabel := widget.NewLabel("")
	from := widget.NewSelect(currencies, nil)
	to := widget.NewSelect(currencies, nil)
	from.SetSelected("THB")
	to.SetSelected("RUB")
	result := widget.NewLabel("Введите сумму")
	result.Wrapping = fyne.TextWrapWord
	update := func(string) {
		value, err := service.ParseDecimal(amount.Text, 2)
		if err != nil {
			result.SetText("Введите сумму с точностью до двух знаков")
			return
		}
		rate := func(currency string) int64 {
			if currency == "RUB" {
				return service.RateScale
			}
			return s.ledger.Rates[currency]
		}
		if from.Selected == to.Selected {
			result.SetText(money(value, to.Selected))
			return
		}
		if strings.TrimSpace(manualRate.Text) != "" {
			rate, err := service.ParseDecimal(manualRate.Text, 6)
			if err != nil || rate <= 0 {
				result.SetText("Введите положительный курс, до шести знаков после запятой")
				return
			}
			converted := value*(rate/service.RateScale) +
				(value*(rate%service.RateScale)+service.RateScale/2)/service.RateScale
			result.SetText("~ " + money(converted, to.Selected))
			return
		}
		source, target := rate(from.Selected), rate(to.Selected)
		if source <= 0 || target <= 0 {
			result.SetText("Введите разовый курс или задайте курсы в настройках")
			return
		}
		result.SetText("~ " + money((value*source+target/2)/target, to.Selected))
	}
	amount.OnChanged = update
	manualRate.OnChanged = update
	changePair := func(string) {
		rateLabel.SetText("Разовый курс: " + to.Selected + " за 1 " + from.Selected)
		manualRate.SetText("")
		update("")
	}
	from.OnChanged = changePair
	to.OnChanged = changePair
	rateLabel.SetText("Разовый курс: " + to.Selected + " за 1 " + from.Selected)
	swap := widget.NewButtonWithIcon("Поменять местами", swapIcon, func() {
		previous := from.Selected
		from.SetSelected(to.Selected)
		to.SetSelected(previous)
	})
	fields := container.NewVBox(widget.NewLabel("Сумма"), amount,
		container.NewGridWithColumns(2, from, to), swap, rateLabel, manualRate, surface(result))
	s.showPage(converterPage, container.NewPadded(sectionTitle("Конвертер валют")), nil,
		container.NewVScroll(container.NewPadded(fields)))
}
