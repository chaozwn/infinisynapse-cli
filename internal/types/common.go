package types

import "encoding/json"

type APIResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type PagerResult struct {
	Items      json.RawMessage `json:"items"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"pageSize"`
	TotalPages int             `json:"totalPages"`
}

type PagerParams struct {
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"pageSize,omitempty"`
	Field    string `json:"field,omitempty"`
	Order    string `json:"order,omitempty"`
}

type TaskListItem struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	TaskName  string `json:"task_name"`
	Status    string `json:"task_status"`
	UpdatedAt string `json:"updatedAt"`
}

type TaskListResponse struct {
	Items []TaskListItem `json:"items"`
	Meta  PagerMeta      `json:"meta"`
}

type PagerMeta struct {
	ItemCount    int `json:"itemCount"`
	TotalItems   int `json:"totalItems"`
	ItemsPerPage int `json:"itemsPerPage"`
	TotalPages   int `json:"totalPages"`
	CurrentPage  int `json:"currentPage"`
}
