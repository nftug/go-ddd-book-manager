package dto

type BookRequest struct {
	Title       string   `json:"title"`
	Authors     []string `json:"authors"`
	ISBN        *string  `json:"isbn"`
	Description *string  `json:"description"`
}
