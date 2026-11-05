package ui

import (
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	fynetest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestCategoryCreatedInUIIsAvailableAndSavesInExpense(t *testing.T) {
	app := fynetest.NewApp()
	defer app.Quit()
	w := app.NewWindow("Копилка")
	storage := openTestStore(t, filepath.Join(t.TempDir(), "ledger.db"))
	l, err := storage.Load()
	if err != nil {
		t.Fatal(err)
	}
	s := screen{window: w, ledger: l, save: storage.Save}
	s.home()
	tapButton(t, w, "Поездки")
	tapButton(t, w, "Новая поездка")
	controls[*widget.Entry](w.Content())[0].SetText("Таиланд")
	tapButton(t, w, "Сохранить")
	journal := s.ledger.Journals[0]
	journal.Opening = map[string]int64{"THB": 100_000}
	s.ledger = s.ledger.WithJournal(journal)
	s.ledger.Rates = map[string]int64{"THB": 2_500_000}
	s.home()

	tapButton(t, w, "Категории")
	tapButton(t, w, "Добавить категорию")
	controls[*widget.Entry](w.Content())[0].SetText("Экскурсии")
	tapButton(t, w, "Путешествие")
	tapButton(t, w, "Сохранить")
	tapButton(t, w, "Настройки")
	tapButton(t, w, "Управление данными")
	exportButtons := 0
	for _, button := range controls[*widget.Button](w.Content()) {
		if button.Text == "Сохранить CSV" || button.Text == "Сохранить JSON" {
			exportButtons++
		}
	}
	if exportButtons != 2 {
		t.Fatalf("expected CSV and JSON export buttons in settings, found %d", exportButtons)
	}
	tapButton(t, w, "Назад")
	tapButton(t, w, "Журнал")
	tapButton(t, w, "Добавить трату")

	choices := controls[*widget.Select](w.Content())
	choices[1].SetSelected("Экскурсии")
	if choices[1].Selected != "Экскурсии" {
		t.Fatalf("category is missing from expense picker: %v", choices[1].Options)
	}
	controls[*widget.Entry](w.Content())[1].SetText("350")
	tapButton(t, w, "Сохранить")
	l, err = storage.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(l.Operations) != 1 || l.CategoryName(l.Operations[0].CategoryID) != "Экскурсии" ||
		l.Category(l.Operations[0].CategoryID).Icon != "trip" ||
		l.Operations[0].Rubles != 87_500 {
		t.Fatalf("category or amount lost on saving expense: %+v", l.Operations)
	}
}

func tapButton(t *testing.T, w fyne.Window, label string) {
	t.Helper()
	for _, tile := range controls[*iconTile](w.Content()) {
		if tile.label == label {
			fynetest.Tap(tile)
			return
		}
	}
	for _, button := range controls[*widget.Button](w.Content()) {
		if button.Text == label {
			fynetest.Tap(button)
			return
		}
	}
	t.Fatalf("button %q not found", label)
}

func controls[T fyne.CanvasObject](root fyne.CanvasObject) []T {
	if value, ok := root.(T); ok {
		return []T{value}
	}
	var result []T
	switch object := root.(type) {
	case *fyne.Container:
		for _, child := range object.Objects {
			result = append(result, controls[T](child)...)
		}
	case *container.Scroll:
		result = append(result, controls[T](object.Content)...)
	}
	return result
}
