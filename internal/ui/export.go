package ui

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"kopilka/internal/service"
)

func (s *screen) exportControls() []fyne.CanvasObject {
	l := s.ledger
	return []fyne.CanvasObject{
		container.NewBorder(nil, nil, iconBadge(theme.DownloadIcon()), nil,
			widget.NewLabelWithStyle("Экспорт", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		widget.NewLabel("CSV — журнал для таблиц"),
		widget.NewLabel("JSON — все данные приложения"),
		widget.NewButtonWithIcon("Сохранить CSV", theme.FileTextIcon(), func() {
			s.exportFile(".csv", func(w io.Writer) error { return writeCSV(w, l) })
		}),
		widget.NewButtonWithIcon("Сохранить JSON", theme.DocumentSaveIcon(), func() {
			s.exportFile(".json", func(w io.Writer) error {
				encoder := json.NewEncoder(w)
				encoder.SetIndent("", "  ")
				return encoder.Encode(l)
			})
		}),
	}
}

func (s *screen) exportFile(extension string, write func(io.Writer) error) {
	picker := dialog.NewFileSave(func(file fyne.URIWriteCloser, err error) {
		if err != nil {
			fyne.Do(func() { dialog.ShowError(err, s.window) })
			return
		}
		if file == nil {
			return
		}
		if err := errors.Join(write(file), file.Close()); err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("не удалось сохранить экспорт: %w", err), s.window)
			})
			return
		}
		fyne.Do(func() { dialog.ShowInformation("Экспорт", "Файл сохранён.", s.window) })
	}, s.window)
	picker.SetFileName("kopilka-" + time.Now().Format(time.DateOnly) + extension)
	picker.SetFilter(storage.NewExtensionFileFilter([]string{extension}))
	picker.Show()
}

func writeCSV(output io.Writer, l service.Ledger) error {
	// BOM and semicolon separator make the file usable in Russian Excel.
	if _, err := io.WriteString(output, "\ufeff"); err != nil {
		return err
	}
	writer := csv.NewWriter(output)
	writer.Comma = ';'
	writer.UseCRLF = true
	if err := writer.Write([]string{
		"id", "date", "type", "category_id", "category", "currency", "amount",
		"received_currency", "received_amount", "rub_per_unit", "rub_equivalent",
		"journal_id", "journal",
	}); err != nil {
		return err
	}
	for _, op := range l.Operations {
		journal, _ := l.Journal(op.JournalID)
		row := []string{
			strconv.FormatInt(op.ID, 10), op.Date, "expense", strconv.FormatInt(op.CategoryID, 10),
			spreadsheetText(l.CategoryName(op.CategoryID)), op.Currency, decimalText(op.Amount, 2),
			"", "", decimalText(op.Rate, 6), decimalText(op.Rubles, 2),
			strconv.FormatInt(op.JournalID, 10), spreadsheetText(journal.Name),
		}
		if op.Received != "" {
			row[2], row[3], row[4] = "exchange", "", ""
			row[7], row[8] = op.Received, decimalText(op.Incoming, 2)
			row[9], row[10] = "", ""
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func spreadsheetText(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed != "" && strings.ContainsRune("=+-@", rune(trimmed[0])) {
		return "'" + value
	}
	return value
}
