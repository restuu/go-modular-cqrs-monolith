package dto

type ProductSearchRequest struct {
	Page         int
	Size         int
	Search       string
	ModelNumbers []string
}
