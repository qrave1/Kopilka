package service

const LedgerVersion = 2

type Ledger struct {
	Version         int              `json:"schema_version"`
	Started         bool             `json:"opening_configured"`
	NextCategoryID  int64            `json:"next_category_id"`
	NextOperationID int64            `json:"next_operation_id"`
	Opening         map[string]int64 `json:"opening"`
	Rates           map[string]int64 `json:"rates"`
	Categories      []Category       `json:"categories"`
	Operations      []Operation      `json:"operations"`
}

func NewLedger() Ledger {
	l := Ledger{Version: LedgerVersion, NextCategoryID: 1, NextOperationID: 1}
	for _, name := range []string{"Еда", "Транспорт", "Жильё", "Развлечения", "Покупки", "Другое"} {
		l.addCategory(name)
	}
	return l
}
