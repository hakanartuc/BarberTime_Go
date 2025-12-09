package utils

type Pagination struct {
	CurrentPage int   `json:"currentPage"`
	TotalPages  int   `json:"totalPages"`
	TotalItems  int64 `json:"totalItems"`
	HasMore     bool  `json:"hasMore"`
}

func NewPagination(page, limit int, total int64) Pagination {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages == 0 {
		totalPages = 1
	}

	return Pagination{
		CurrentPage: page,
		TotalPages:  totalPages,
		TotalItems:  total,
		HasMore:     page < totalPages,
	}
}

