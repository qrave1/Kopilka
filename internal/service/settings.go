package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const RateScale = 1_000_000

func Currencies() []string {
	return []string{"RUB", "USD", "THB"}
}

type Balance struct {
	Currency string
	Amount   int64
}

func (l Ledger) WorsenedNegativeBalances(journalID int64, next Ledger) []Balance {
	before, after := l.Balances(journalID), next.Balances(journalID)
	var warnings []Balance
	for _, currency := range Currencies() {
		if after[currency] < 0 && after[currency] < before[currency] {
			warnings = append(warnings, Balance{Currency: currency, Amount: after[currency]})
		}
	}
	return warnings
}

func (l Ledger) SuggestedExchange(journal Journal) (from, to string) {
	balance := l.Balances(journal.ID)
	if journal.Currency == "THB" {
		if balance["USD"] > 0 && balance["THB"] <= 0 {
			return "USD", "THB"
		}
		return "RUB", "USD"
	}
	if journal.Currency == "RUB" {
		return "USD", "RUB"
	}
	return "RUB", "USD"
}

func (j Journal) WithOpeningCash(inputs map[string]string) (Journal, error) {
	j.Opening = make(map[string]int64, len(inputs))
	for _, currency := range Currencies() {
		input := inputs[currency]
		if strings.TrimSpace(input) == "" {
			input = "0"
		}
		value, err := ParseDecimal(input, 2)
		if err != nil {
			return Journal{}, fmt.Errorf("проверь начальный остаток %s", currency)
		}
		j.Opening[currency] = value
	}
	j.Started = true
	return j, nil
}

func (l Ledger) WithRates(inputs map[string]string) (Ledger, error) {
	l.Rates = make(map[string]int64, len(inputs))
	for _, currency := range []string{"THB", "USD"} {
		value, err := ReadRate(inputs[currency])
		if err != nil {
			return Ledger{}, fmt.Errorf("%s: %w", currency, err)
		}
		l.Rates[currency] = value
	}
	return l, nil
}

func (l Ledger) Balances(journalID int64) map[string]int64 {
	journal, _ := l.Journal(journalID)
	balances := make(map[string]int64, len(journal.Opening))
	for currency, amount := range journal.Opening {
		balances[currency] = amount
	}
	for _, op := range l.Operations {
		if op.JournalID != journalID {
			continue
		}
		balances[op.Currency] -= op.Amount
		if op.Received != "" {
			balances[op.Received] += op.Incoming
		}
	}
	return balances
}

// Seven whole digits keep amount times rate within int64 at the maximum rate.
func ParseDecimal(input string, places int) (int64, error) {
	whole, fraction, _ := strings.Cut(strings.ReplaceAll(strings.TrimSpace(input), ",", "."), ".")
	if whole == "" && fraction != "" {
		whole = "0"
	}
	if len(whole) == 0 || len(whole) > 7 || len(fraction) > places {
		return 0, errors.New("некорректное число")
	}
	digits := whole + fraction + strings.Repeat("0", places-len(fraction))
	for _, digit := range digits {
		if digit < '0' || digit > '9' {
			return 0, errors.New("некорректное число")
		}
	}
	return strconv.ParseInt(digits, 10, 64)
}

func ReadRate(input string) (int64, error) {
	if strings.TrimSpace(input) == "" {
		return 0, nil
	}
	rate, err := ParseDecimal(input, 6)
	if err != nil || rate <= 0 || rate > 1_000*RateScale {
		return 0, errors.New("курс должен быть больше нуля и не превышать 1 000 ₽, до 6 знаков после запятой")
	}
	return rate, nil
}
