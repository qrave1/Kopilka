package ui

import (
	"bytes"
	"encoding/csv"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"kopilka/internal/service"
	"kopilka/internal/store"
)

func TestTripJournalPersistsBalancesAndHistoricalRate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ledger.db")
	storage := openTestStore(t, path)
	l, err := storage.Load()
	if err != nil {
		t.Fatal(err)
	}
	l = addTestTrip(t, l)
	journal := l.Journals[0]
	journal.Opening = map[string]int64{"RUB": 5_000_000}
	l = l.WithJournal(journal)
	l.Rates = map[string]int64{"THB": 2_500_000}
	l, err = l.WithCategoryName(0, "Экскурсии")
	if err != nil {
		t.Fatal(err)
	}
	_, categories := l.ExpenseCategories(0)
	id := categories["Экскурсии"]
	l = l.WithOperation(-1, service.Operation{JournalID: 1, Date: "2026-11-01", Currency: "RUB", Amount: 3_000_000, Received: "USD", Incoming: 30_000})
	if balance := l.Balances(1); balance["USD"] != 30_000 || balance["THB"] != 0 {
		t.Fatalf("Balances before exchanging USD = %v", balance)
	}
	l = l.WithOperation(-1, service.Operation{JournalID: 1, Date: "2026-11-02", Currency: "USD", Amount: 20_000, Received: "THB", Incoming: 660_000})
	expense, err := l.Expense("2026-11-03", "THB", id, 35_000, service.Operation{})
	if err != nil || expense.Rubles != 87_500 {
		t.Fatalf("initial expense = %+v, error = %v", expense, err)
	}
	expense.JournalID = 1
	l = l.WithOperation(-1, expense)
	if err := storage.Save(l); err != nil {
		t.Fatal(err)
	}
	l.Rates["THB"] = 3_000_000
	edited, err := l.Expense(expense.Date, "THB", expense.CategoryID, 40_000, expense)
	if err != nil || edited.Rubles != 100_000 || edited.Rate != expense.Rate {
		t.Fatalf("edited expense = %+v, error = %v", edited, err)
	}
	fresh, err := l.Expense(expense.Date, "THB", expense.CategoryID, 40_000, service.Operation{})
	if err != nil || fresh.Rubles != 120_000 {
		t.Fatalf("new expense = %+v, error = %v", fresh, err)
	}
	edited.JournalID = 1
	l = l.WithOperation(2, edited)
	l, err = l.WithTrip("Вторая поездка", "USD")
	if err != nil {
		t.Fatal(err)
	}
	second, err := l.Journals[1].WithOpeningCash(map[string]string{"USD": "100"})
	if err != nil {
		t.Fatal(err)
	}
	l = l.WithJournal(second).WithOperation(-1, service.Operation{
		JournalID: second.ID, Date: "2026-11-04", Currency: "USD", Amount: 1_000,
		CategoryID: id, Rate: 90 * service.RateScale, Rubles: 90_000,
	})
	if err := storage.Save(l); err != nil {
		t.Fatal(err)
	}
	if err := storage.Close(); err != nil {
		t.Fatal(err)
	}
	storage = openTestStore(t, path)
	got, err := storage.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, l) {
		t.Fatalf("reopened ledger = %+v, want %+v", got, l)
	}
	if total := got.TotalSpentRubles(1); total != 100_000 {
		t.Fatalf("total spent = %d, want 100000 kopecks", total)
	}
	if total, balance := got.TotalSpentRubles(second.ID), got.Balances(second.ID)["USD"]; total != 90_000 || balance != 9_000 {
		t.Fatalf("second trip total = %d, USD balance = %d", total, balance)
	}
	want := map[string]int64{"RUB": 2_000_000, "USD": 10_000, "THB": 620_000}
	if Balances := got.Balances(1); !reflect.DeepEqual(Balances, want) {
		t.Fatalf("Balances = %v, want %v", Balances, want)
	}
	if balance := got.WithoutOperation(1).Balances(1)["THB"]; balance != -40_000 {
		t.Fatalf("THB after deleting exchange = %d, want -40000", balance)
	}
	invalid := edited
	invalid.CategoryID = 9999
	if err := storage.Save(got.WithOperation(2, invalid)); err == nil {
		t.Fatal("expected foreign-key error for missing category")
	}
	afterFailure, err := storage.Load()
	if err != nil || !reflect.DeepEqual(afterFailure, got) {
		t.Fatalf("failed save changed the ledger: %+v, err = %v", afterFailure, err)
	}
}

func TestExpenseRoundsDecimalInputToKopecks(t *testing.T) {
	amount, err := service.ParseDecimal("350,50", 2)
	if err != nil || amount != 35_050 {
		t.Fatalf("amount = %d, error = %v", amount, err)
	}
	rate, err := service.ReadRate("2.5001")
	if err != nil {
		t.Fatal(err)
	}
	l := service.Ledger{Rates: map[string]int64{"THB": rate}}
	expense, err := l.Expense("2026-11-03", "THB", 1, amount, service.Operation{})
	if err != nil || expense.Rubles != 87_629 {
		t.Fatalf("expense = %+v, error = %v", expense, err)
	}
}

func TestArchivedCategoryAndExport(t *testing.T) {
	storage := openTestStore(t, filepath.Join(t.TempDir(), "ledger.db"))
	l, err := storage.Load()
	if err != nil {
		t.Fatal(err)
	}
	l = addTestTrip(t, l)
	l.Categories = l.Categories[:1]
	l = l.WithOperation(-1, service.Operation{JournalID: 1, Date: "2026-11-01", Currency: "RUB", Amount: 3_000_000, Received: "USD", Incoming: 30_000})
	op := service.Operation{JournalID: 1, Date: "2026-11-03", Currency: "THB", Amount: 35_000,
		CategoryID: l.Categories[0].ID, Rate: 2_500_000, Rubles: 87_500}
	l = l.WithOperation(-1, op)
	l, err = l.WithCategoryName(op.CategoryID, "Кафе; ужин")
	if err != nil {
		t.Fatal(err)
	}
	l = l.WithoutCategory(0)
	if err := storage.Save(l); err != nil {
		t.Fatal(err)
	}
	l, err = storage.Load()
	if err != nil {
		t.Fatal(err)
	}
	if names, _ := l.ExpenseCategories(0); len(names) != 0 {
		t.Fatalf("archived category available for new expenses: %v", names)
	}
	if names, _ := l.ExpenseCategories(op.CategoryID); len(names) != 1 || names[0] != "Кафе; ужин" {
		t.Fatalf("archived category unavailable for existing expense: %v", names)
	}
	var output bytes.Buffer
	if err := writeCSV(&output, l); err != nil {
		t.Fatal(err)
	}
	reader := csv.NewReader(strings.NewReader(strings.TrimPrefix(output.String(), "\ufeff")))
	reader.Comma = ';'
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if rows[1][2] != "exchange" || rows[1][7] != "USD" || rows[1][10] != "" ||
		rows[2][4] != "Кафе; ужин" || rows[2][9] != "2,5" || rows[2][10] != "875" || rows[2][12] != "Таиланд" {
		t.Fatalf("unexpected export: %v", rows)
	}
}

func openTestStore(t *testing.T, path string) *store.Store {
	t.Helper()
	storage, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { storage.Close() })
	return storage
}

func addTestTrip(t *testing.T, l service.Ledger) service.Ledger {
	t.Helper()
	next, err := l.WithTrip("Таиланд", "THB")
	if err != nil {
		t.Fatal(err)
	}
	return next
}
