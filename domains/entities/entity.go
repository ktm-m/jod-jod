package entities

type Pagination struct {
	PageItem int `query:"page-item"`
	Page     int `query:"page"`
}
