package types

type RelatedRag struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled int    `json:"enabled"`
}

type DatabaseItem struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Type        string       `json:"type"`
	Description string       `json:"description"`
	Source      string       `json:"source"`
	Enabled     int          `json:"enabled"`
	RagList     []RelatedRag `json:"rag_list"`
	CreatedAt   string       `json:"createdAt"`
	UpdatedAt   string       `json:"updatedAt"`
}

type DatabaseListResponse struct {
	Items []DatabaseItem `json:"items"`
	Meta  PagerMeta      `json:"meta"`
}

type DatabaseEnabledRequest struct {
	IDs     []string `json:"ids"`
	Enabled int      `json:"enabled"`
}
