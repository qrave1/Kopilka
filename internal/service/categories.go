package service

import (
	"errors"
	"slices"
	"strings"
	"unicode/utf8"
)

type Category struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Archived bool   `json:"archived,omitempty"`
}

func (l *Ledger) addCategory(name string) int64 {
	id := l.NextCategoryID
	l.NextCategoryID++
	l.Categories = append(l.Categories, Category{ID: id, Name: name})
	return id
}

func (l Ledger) CategoryName(id int64) string {
	for _, c := range l.Categories {
		if c.ID == id {
			return c.Name
		}
	}
	return ""
}

func (l Ledger) CategoryInUse(id int64) bool {
	return slices.ContainsFunc(l.Operations, func(op Operation) bool { return op.CategoryID == id })
}

func (l Ledger) ExpenseCategories(selectedID int64) ([]string, map[string]int64) {
	names := make([]string, 0, len(l.Categories))
	ids := make(map[string]int64, len(l.Categories))
	for _, c := range l.Categories {
		if c.Archived && c.ID != selectedID {
			continue
		}
		names = append(names, c.Name)
		ids[c.Name] = c.ID
	}
	return names, ids
}

func (l Ledger) WithCategoryName(id int64, name string) (Ledger, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 40 || strings.ContainsAny(name, "\r\n\t") {
		return Ledger{}, errors.New("введи название категории в одну строку, до 40 символов")
	}
	for _, c := range l.Categories {
		if c.ID != id && strings.EqualFold(c.Name, name) {
			return Ledger{}, errors.New("такая категория уже есть, проверь и архив")
		}
	}
	l.Categories = slices.Clone(l.Categories)
	if id == 0 {
		l.addCategory(name)
	} else {
		index := slices.IndexFunc(l.Categories, func(c Category) bool { return c.ID == id })
		l.Categories[index].Name = name
	}
	return l, nil
}

func (l Ledger) WithoutCategory(index int) Ledger {
	l.Categories = slices.Clone(l.Categories)
	if l.CategoryInUse(l.Categories[index].ID) {
		l.Categories[index].Archived = true
	} else {
		l.Categories = slices.Delete(l.Categories, index, index+1)
	}
	return l
}

func (l Ledger) RestoreCategory(index int) Ledger {
	l.Categories = slices.Clone(l.Categories)
	l.Categories[index].Archived = false
	return l
}
