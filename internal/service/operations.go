package service

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type Operation struct {
	ID         int64  `json:"id"`
	Date       string `json:"date"`
	Currency   string `json:"currency"`
	Amount     int64  `json:"amount"`
	Received   string `json:"received_currency,omitempty"`
	Incoming   int64  `json:"received_amount,omitempty"`
	CategoryID int64  `json:"category_id,omitempty"`
	Rate       int64  `json:"rate,omitempty"`
	Rubles     int64  `json:"rubles,omitempty"`
}

type OperationInput struct {
	Date, Currency, Amount string
	Received, Incoming     string
	CategoryID             int64
	Exchange               bool
}

func (l Ledger) PrepareOperation(input OperationInput, previous Operation) (Operation, error) {
	day := strings.TrimSpace(input.Date)
	if _, err := time.Parse(time.DateOnly, day); err != nil {
		return Operation{}, errors.New("введи дату в формате ГГГГ-ММ-ДД")
	}
	amount, err := ParseDecimal(input.Amount, 2)
	if err != nil || amount == 0 {
		return Operation{}, errors.New("введи положительную сумму с точностью до двух знаков")
	}
	if input.Exchange {
		incoming, err := ParseDecimal(input.Incoming, 2)
		if err != nil || incoming == 0 || input.Received == input.Currency {
			return Operation{}, errors.New("выбери разные валюты и укажи положительную сумму получения")
		}
		return Operation{
			Date: day, Currency: input.Currency, Amount: amount,
			Received: input.Received, Incoming: incoming,
		}, nil
	}
	if input.CategoryID == 0 {
		return Operation{}, errors.New("выбери категорию")
	}
	return l.Expense(day, input.Currency, input.CategoryID, amount, previous)
}

// TotalSpentRubles returns the sum of saved expense equivalents in kopecks.
func (l Ledger) TotalSpentRubles() int64 {
	var total int64
	for _, op := range l.Operations {
		if op.Received == "" {
			total += op.Rubles
		}
	}
	return total
}

func (l Ledger) WithOperation(index int, op Operation) Ledger {
	operations := append([]Operation(nil), l.Operations...)
	if index < 0 {
		op.ID = l.NextOperationID
		l.NextOperationID++
		operations = append(operations, op)
	} else {
		op.ID = operations[index].ID
		operations[index] = op
	}
	l.Operations = operations
	return l
}

func (l Ledger) WithoutOperation(index int) Ledger {
	operations := append([]Operation(nil), l.Operations[:index]...)
	l.Operations = append(operations, l.Operations[index+1:]...)
	return l
}

func (l Ledger) Expense(date, currency string, categoryID, amount int64, previous Operation) (Operation, error) {
	rate := l.Rates[currency]
	if currency == previous.Currency {
		rate = previous.Rate
	}
	if currency == "RUB" {
		rate = RateScale
	}
	if rate <= 0 {
		return Operation{}, fmt.Errorf("укажите курс %s в настройках", currency)
	}
	return Operation{
		Date: date, Currency: currency, CategoryID: categoryID, Amount: amount,
		Rate: rate, Rubles: (amount*rate + RateScale/2) / RateScale,
	}, nil
}
