package service

const LedgerVersion = 3

type Ledger struct {
	Version         int              `json:"schema_version"`
	NextCategoryID  int64            `json:"next_category_id"`
	NextOperationID int64            `json:"next_operation_id"`
	Journals        []Journal        `json:"journals"`
	Rates           map[string]int64 `json:"rates"`
	Categories      []Category       `json:"categories"`
	Operations      []Operation      `json:"operations"`
}
