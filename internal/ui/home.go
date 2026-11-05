package ui

import (
	"slices"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"kopilka/internal/service"
)

func (s *screen) home() {
	balance := s.ledger.Balances()
	summary := widget.NewCard("", "", container.NewVBox(
		container.NewGridWithColumns(2,
			summaryMetric("Остаток батов", moneyAmount(balance["THB"]), fyne.TextAlignLeading),
			summaryMetric("Всего потрачено", money(s.ledger.TotalSpentRubles(), "RUB"), fyne.TextAlignTrailing),
		),
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			alignedLabel(money(balance["RUB"], "RUB"), fyne.TextAlignLeading, false),
			alignedLabel(money(balance["USD"], "USD"), fyne.TextAlignTrailing, false),
		),
	))
	add := widget.NewButtonWithIcon("Добавить трату", theme.ContentAddIcon(), s.afterTap(func() { s.editOperation(-1, false) }))
	add.Importance = widget.HighImportance
	exchangeLabel := "Обмен"
	from, _ := s.ledger.SuggestedExchange()
	switch {
	case from == "USD":
		exchangeLabel = "Обмен $ в THB"
	case balance["RUB"] > 0 && balance["USD"] == 0 && balance["THB"] == 0:
		exchangeLabel = "Обмен ₽ в $"
	}
	exchange := widget.NewButtonWithIcon(exchangeLabel, theme.ViewRefreshIcon(), s.afterTap(func() { s.editOperation(-1, true) }))
	content := container.NewVBox(
		widget.NewRichTextFromMarkdown("## Копилка"),
		summary,
		container.NewGridWithColumns(2, add, exchange),
	)
	if !s.ledger.Started {
		content.Add(widget.NewButton("Указать начальные наличные", s.afterTap(s.openingCash)))
	}
	content.Add(widget.NewLabelWithStyle("Операции", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	header := container.NewPadded(content)

	indices := make([]int, len(s.ledger.Operations))
	for i := range indices {
		indices[i] = len(indices) - i - 1
	}
	slices.SortStableFunc(indices, func(a, b int) int {
		return strings.Compare(s.ledger.Operations[b].Date, s.ledger.Operations[a].Date)
	})
	if len(indices) == 0 {
		message := widget.NewLabel("Здесь появятся твои траты и обмены.")
		message.Wrapping = fyne.TextWrapWord
		s.showPage(journalPage, header, nil, container.NewPadded(message))
		return
	}
	list := widget.NewList(
		func() int { return len(indices) },
		func() fyne.CanvasObject { return newJournalRow() },
		func(id widget.ListItemID, item fyne.CanvasObject) {
			op := s.ledger.Operations[indices[id]]
			item.(*journalRow).update(op, s.ledger.CategoryName(op.CategoryID))
		},
	)
	list.OnSelected = func(id widget.ListItemID) {
		index := indices[id]
		list.Unselect(id)
		s.editOperation(index, s.ledger.Operations[index].Received != "")
	}
	s.showPage(journalPage, header, nil, container.NewPadded(list))
}

func alignedLabel(text string, align fyne.TextAlign, bold bool) *widget.Label {
	label := widget.NewLabelWithStyle(text, align, fyne.TextStyle{Bold: bold})
	label.Truncation = fyne.TextTruncateEllipsis
	return label
}

func summaryMetric(title, amount string, align fyne.TextAlign) *fyne.Container {
	value := widget.NewRichText(&widget.TextSegment{
		Text: amount,
		Style: widget.RichTextStyle{
			Alignment: align,
			SizeName:  theme.SizeNameHeadingText,
			TextStyle: fyne.TextStyle{Bold: true},
		},
	})
	value.Truncation = fyne.TextTruncateEllipsis
	return container.NewVBox(alignedLabel(title, align, false), value)
}

type journalRow struct {
	widget.Card
	title, date, amount, detail *widget.Label
}

func newJournalRow() *journalRow {
	row := &journalRow{
		title:  alignedLabel("Категория", fyne.TextAlignLeading, true),
		date:   alignedLabel("Дата", fyne.TextAlignLeading, false),
		amount: alignedLabel("0,00 ₽", fyne.TextAlignTrailing, true),
		detail: alignedLabel(" ", fyne.TextAlignTrailing, false),
	}
	row.Content = container.NewGridWithColumns(2,
		container.NewVBox(row.title, row.date),
		container.NewVBox(row.amount, row.detail),
	)
	row.ExtendBaseWidget(row)
	return row
}

func (row *journalRow) update(op service.Operation, category string) {
	detail := " "
	switch {
	case op.Received != "":
		category = "Обмен"
		detail = "+ " + money(op.Incoming, op.Received)
	case op.Currency != "RUB":
		detail = "~ " + money(op.Rubles, "RUB")
	}
	row.title.SetText(category)
	row.date.SetText(op.Date)
	row.amount.SetText("- " + money(op.Amount, op.Currency))
	row.detail.SetText(detail)
}
