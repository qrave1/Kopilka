package main

import (
	"path/filepath"

	"fyne.io/fyne/v2/app"

	"kopilka/internal/service"
	"kopilka/internal/store"
	"kopilka/internal/ui"
)

const appID = "io.kopilka.travel"

func main() {
	a := app.NewWithID(appID)
	storage, err := store.Open(filepath.Join(a.Storage().RootURI().Path(), "ledger.db"))
	if err != nil {
		ui.Run(a, service.Ledger{}, nil, err)
		return
	}
	defer storage.Close()

	ledger, err := storage.Load()
	ui.Run(a, ledger, storage.Save, err)
}
