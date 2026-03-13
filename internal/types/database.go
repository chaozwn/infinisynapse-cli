package types

type DatabaseQueryParams struct {
	PagerParams
	Name            string `json:"name,omitempty"`
	Type            string `json:"type,omitempty"`
	Enabled         *int   `json:"enabled,omitempty"`
	SubscribeSource string `json:"subscribeSource,omitempty"`
	Source          string `json:"source,omitempty"`
}

type DatabaseAddParams struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Config      string `json:"config"`
	Enabled     int    `json:"enabled"`
	Description string `json:"description,omitempty"`
}

type DatabaseDeleteParams struct {
	IDs []string `json:"ids"`
}

type DatabaseEnabledParams struct {
	IDs     []string `json:"ids"`
	Enabled int      `json:"enabled"`
}

type DatabaseEditParams struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Config      string `json:"config"`
	Enabled     int    `json:"enabled"`
	Description string `json:"description"`
}

type DatabaseTestConnectionParams struct {
	Type   string `json:"type"`
	Config string `json:"config"`
}

type DatabaseBindRagParams struct {
	DatabaseID string   `json:"databaseId"`
	RagIDs     []string `json:"ragIds"`
}
