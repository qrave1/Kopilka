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
	journal, ok := s.ledger.Journal(s.journalID)
	if !ok {
		for _, candidate := range s.ledger.Journals {
			if candidate.Kind == service.TripJournal && !candidate.Archived {
				journal, ok = candidate, true
				s.journalID = candidate.ID
				break
			}
		}
	}
	if !ok {
		s.trips()
		return
	}
	balance := s.ledger.Balances(journal.ID)
	balanceTitle := map[string]string{"RUB": "Остаток рублей", "USD": "Остаток долларов", "THB": "Остаток батов"}
	var otherBalances []fyne.CanvasObject
	for _, currency := range currencies {
		if currency != journal.Currency {
			otherBalances = append(otherBalances, alignedLabel(money(balance[currency], currency), fyne.TextAlignCenter, false))
		}
	}
	summary := container.NewGridWithColumns(2,
		surface(container.NewVBox(iconBadge(walletIcon),
			summaryMetric(balanceTitle[journal.Currency], moneyAmount(balance[journal.Currency]), fyne.TextAlignLeading))),
		surface(container.NewVBox(iconBadge(theme.HistoryIcon()),
			summaryMetric("Всего потрачено", money(s.ledger.TotalSpentRubles(journal.ID), "RUB"), fyne.TextAlignLeading))),
	)
	title := container.NewBorder(nil, nil, iconBadge(suitcaseIcon), nil, sectionTitle(journal.Name))
	content := container.NewVBox(title, summary, container.NewGridWithColumns(2, otherBalances...))
	if journal.Archived {
		content.Add(widget.NewLabel("Поездка завершена"))
	} else {
		content.Add(surface(container.NewGridWithColumns(3,
			newIconTile("Добавить трату", theme.ContentAddIcon(), true, s.afterTap(func() { s.editOperation(-1, false) })),
			newIconTile("Обмен", swapIcon, false, s.afterTap(func() { s.editOperation(-1, true) })),
			newIconTile("Наличные", walletIcon, false, s.afterTap(s.openingCash)),
		)))
	}
	content.Add(widget.NewLabelWithStyle("Операции", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	header := container.NewPadded(content)

	var indices []int
	for i := len(s.ledger.Operations) - 1; i >= 0; i-- {
		if s.ledger.Operations[i].JournalID == journal.ID {
			indices = append(indices, i)
		}
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
		if !journal.Archived {
			s.editOperation(index, s.ledger.Operations[index].Received != "")
		}
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
	widget.BaseWidget
	glyph                       *widget.Icon
	title, date, amount, detail *widget.Label
}

func newJournalRow() *journalRow {
	row := &journalRow{
		glyph:  widget.NewIcon(tagIcon),
		title:  alignedLabel("Категория", fyne.TextAlignLeading, true),
		date:   alignedLabel("Дата", fyne.TextAlignLeading, false),
		amount: alignedLabel("0,00 ₽", fyne.TextAlignTrailing, true),
		detail: alignedLabel(" ", fyne.TextAlignTrailing, false),
	}
	row.ExtendBaseWidget(row)
	return row
}

func (row *journalRow) CreateRenderer() fyne.WidgetRenderer {
	content := container.NewGridWithColumns(2,
		container.NewBorder(nil, nil, container.NewPadded(container.NewCenter(row.glyph)), nil,
			container.NewVBox(row.title, row.date)),
		container.NewVBox(row.amount, row.detail),
	)
	return widget.NewSimpleRenderer(surface(content))
}

func (row *journalRow) update(op service.Operation, category string) {
	detail := " "
	icon := categoryIcon(category)
	switch {
	case op.Received != "":
		category = "Обмен"
		icon = swapIcon
		detail = "+ " + money(op.Incoming, op.Received)
	case op.Currency != "RUB":
		detail = "~ " + money(op.Rubles, "RUB")
	}
	row.title.SetText(category)
	row.glyph.SetResource(theme.NewPrimaryThemedResource(icon))
	row.date.SetText(op.Date)
	row.amount.SetText("- " + money(op.Amount, op.Currency))
	row.detail.SetText(detail)
}
