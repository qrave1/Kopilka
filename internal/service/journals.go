package service

import (
	"errors"
	"slices"
	"strings"
	"unicode/utf8"
)

type JournalKind string

const (
	TripJournal     JournalKind = "trip"
	EverydayJournal JournalKind = "everyday"
)

type Journal struct {
	ID       int64            `json:"id"`
	Kind     JournalKind      `json:"kind"`
	Name     string           `json:"name"`
	Currency string           `json:"currency"`
	Archived bool             `json:"archived,omitempty"`
	Started  bool             `json:"opening_configured"`
	Opening  map[string]int64 `json:"opening"`
}

func (l Ledger) WithTrip(name, currency string) (Ledger, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 40 || strings.ContainsAny(name, "\r\n\t") {
		return Ledger{}, errors.New("введи название поездки в одну строку, до 40 символов")
	}
	if !slices.Contains(Currencies(), currency) {
		return Ledger{}, errors.New("выбери валюту поездки")
	}
	for _, journal := range l.Journals {
		if strings.EqualFold(journal.Name, name) {
			return Ledger{}, errors.New("поездка с таким названием уже есть")
		}
	}
	var id int64 = 1
	for _, journal := range l.Journals {
		id = max(id, journal.ID+1)
	}
	l.Journals = append(slices.Clone(l.Journals), Journal{ID: id, Kind: TripJournal, Name: name, Currency: currency})
	return l, nil
}

func (l Ledger) Journal(id int64) (Journal, bool) {
	for _, j := range l.Journals {
		if j.ID == id {
			return j, true
		}
	}
	return Journal{}, false
}

func (l Ledger) WithJournal(journal Journal) Ledger {
	l.Journals = slices.Clone(l.Journals)
	for i, current := range l.Journals {
		if current.ID == journal.ID {
			l.Journals[i] = journal
			break
		}
	}
	return l
}
