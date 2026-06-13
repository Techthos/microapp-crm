package models

// PipelineSummary is the read-only funnel + pipeline aggregate (UC-18). Deal
// values are grouped by currency and never summed across currencies (the app is
// offline and does no FX conversion — see spec Non-Goals).
type PipelineSummary struct {
	DealsByStage  []StageSummary `json:"dealsByStage"`
	LeadsByStatus []StatusCount  `json:"leadsByStatus"`
}

// StageSummary is the deal count and per-currency value totals for one stage.
type StageSummary struct {
	Stage  DealStage       `json:"stage"`
	Count  int             `json:"count"`
	Totals []CurrencyTotal `json:"totals"`
}

// CurrencyTotal is a summed value for a single currency code.
type CurrencyTotal struct {
	Currency string  `json:"currency"`
	Total    float64 `json:"total"`
}

// StatusCount is the number of leads in one funnel status.
type StatusCount struct {
	Status LeadStatus `json:"status"`
	Count  int        `json:"count"`
}
