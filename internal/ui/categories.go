package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"kopilka/internal/service"
)

func (s *screen) categories() {
	header := container.NewVBox(
		widget.NewLabel("Категории"),
		widget.NewButton("Добавить категорию", s.afterTap(func() { s.editCategory(-1) })),
	)
	active := container.NewVBox()
	archive := container.NewVBox()
	for i, c := range s.ledger.Categories {
		name := widget.NewLabel(c.Name)
		name.Truncation = fyne.TextTruncateEllipsis
		row := container.NewBorder(nil, nil, nil,
			widget.NewButton("Изменить", s.afterTap(func() { s.editCategory(i) })), name)
		if c.Archived {
			archive.Add(row)
		} else {
			active.Add(row)
		}
	}
	if len(active.Objects) == 0 {
		active.Add(widget.NewLabel("Добавь категорию для новой траты."))
	}
	if len(archive.Objects) > 0 {
		active.Add(widget.NewSeparator())
		active.Add(widget.NewLabel("Архив"))
		active.Add(archive)
	}
	s.showPage(categoriesPage, header, nil, container.NewVScroll(active))
}

func (s *screen) editCategory(index int) {
	var current service.Category
	if index >= 0 {
		current = s.ledger.Categories[index]
	}
	name := entry(current.Name, "Название категории")
	fields := container.NewVBox(name)
	if current.Archived {
		fields.Add(widget.NewButton("Восстановить категорию", func() {
			next := s.ledger.RestoreCategory(index)
			if s.persist(next) {
				s.categories()
			}
		}))
	}
	var remove func()
	if index >= 0 && (!current.Archived || !s.ledger.CategoryInUse(current.ID)) {
		remove = func() {
			message := "Категория будет удалена."
			if s.ledger.CategoryInUse(current.ID) {
				message = "Категория останется в старых тратах и будет скрыта для новых. Её можно восстановить из архива."
			}
			dialog.ShowConfirm("Удалить категорию?", message, func(ok bool) {
				if ok && s.persist(s.ledger.WithoutCategory(index)) {
					s.categories()
				}
			}, s.window)
		}
	}
	s.form("Категория", fields, s.categories, func() {
		next, err := s.ledger.WithCategoryName(current.ID, name.Text)
		if err != nil {
			dialog.ShowError(err, s.window)
			return
		}
		if s.persist(next) {
			s.categories()
		}
	}, remove)
}
