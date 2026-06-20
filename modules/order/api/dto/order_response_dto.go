package dto

type OrderDetailResponse struct {
	OrderNumber string
	Items       []OrderItemResponse
}

type OrderItemResponse struct {
}

type OrderListItemResponse struct {
	Search string
	Page   int
	Size   int
}

type OrderSearchResponse struct {
	Items []OrderListItemResponse
	Total int
}
