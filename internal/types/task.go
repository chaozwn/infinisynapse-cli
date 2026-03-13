package types

type TaskQueryParams struct {
	PagerParams
	TaskName             string `json:"task_name,omitempty"`
	CategoryName         string `json:"category_name,omitempty"`
	CategoryID           string `json:"category_id,omitempty"`
	IsInRag              string `json:"is_in_rag,omitempty"`
	VirtualEchartCategory string `json:"virtual_echart_category,omitempty"`
}

type TaskDeleteParams struct {
	IDs []string `json:"ids"`
}

type TaskCategoryAddParams struct {
	CategoryName string `json:"category_name"`
}

type TaskCategoryUpdateParams struct {
	ID           string `json:"id"`
	CategoryName string `json:"category_name"`
}

type TaskCategoryDeleteParams struct {
	IDs []string `json:"ids"`
}

type TaskCategoryQueryParams struct {
	PagerParams
	CategoryName string `json:"category_name,omitempty"`
}

type TaskRerunSQLParams struct {
	ID        string `json:"id"`
	ChatIndex int    `json:"chat_index"`
}

type TaskSaveToRagParams struct {
	TaskID string `json:"taskId"`
	Action string `json:"action"`
}

type TaskPreviewFileParams struct {
	TaskID   string `json:"taskId"`
	FileName string `json:"fileName"`
}
