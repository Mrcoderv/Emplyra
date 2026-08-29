package utils

import (
	"math"
	"strconv"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Offset   int `json:"-"`
	Limit    int `json:"-"`
}

func NewPagination(pageStr, sizeStr string) Pagination {
	page := atoi(pageStr, 1)
	size := atoi(sizeStr, DefaultPageSize)
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = DefaultPageSize
	}
	if size > MaxPageSize {
		size = MaxPageSize
	}
	return Pagination{Page: page, PageSize: size, Offset: (page - 1) * size, Limit: size}
}

func (p Pagination) TotalPages(total int64) int {
	if total == 0 {
		return 0
	}
	tp := int(math.Ceil(float64(total) / float64(p.PageSize)))
	if tp < 1 {
		tp = 1
	}
	return tp
}

func atoi(s string, def int) int {
	if s == "" {
		return def
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}

func CloneString(s string) *string {
	return &s
}

type PageQuery struct {
	Page     string `form:"page"`
	PageSize string `form:"page_size"`
}
