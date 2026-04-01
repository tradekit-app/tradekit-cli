package types

// Response is the standard API response envelope from the TradeKit gateway.
type Response[T any] struct {
	Success bool       `json:"success"`
	Data    T          `json:"data,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
	Meta    *Meta      `json:"meta,omitempty"`
}

type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"perPage,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"totalPages,omitempty"`
}
