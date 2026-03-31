package types

type RelatedDatabase struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled int    `json:"enabled"`
}

type RagItem struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Source       string            `json:"source"`
	Enabled      int               `json:"enabled"`
	DatabaseList []RelatedDatabase `json:"database_list"`
	CreatedAt    string            `json:"createdAt"`
	UpdatedAt    string            `json:"updatedAt"`
}

type RagListResponse struct {
	Items []RagItem `json:"items"`
	Meta  PagerMeta `json:"meta"`
}

type RagEnabledRequest struct {
	IDs     []string `json:"ids"`
	Enabled int      `json:"enabled"`
}
