package types

type RagItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
	Enabled     int    `json:"enabled"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type RagListResponse struct {
	Items []RagItem `json:"items"`
	Meta  PagerMeta `json:"meta"`
}

type RagEnabledRequest struct {
	IDs     []string `json:"ids"`
	Enabled int      `json:"enabled"`
}
