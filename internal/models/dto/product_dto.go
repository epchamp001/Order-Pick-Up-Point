package dto

import "time"

// ProductDTO godoc
// @Description Represents a product with its details.
type ProductDTO struct {
	Id          string    `json:"id,omitempty" example:"prod123"`
	DateTime    time.Time `json:"dateTime,omitempty" example:"2025-04-09T15:04:05Z"`
	Type        string    `json:"type" example:"electronics"`
	ReceptionId string    `json:"receptionId" example:"recv456"`
}

// ProductsPostRequest godoc
// @Description Request payload for adding a product to a reception.
type ProductsPostRequest struct {
	Type  string `json:"type" binding:"required" example:"clothes"`
	PvzId string `json:"pvzId" binding:"required" example:"pvz789"`
}

// ProductsPostResponse godoc
// @Description Response returned after successfully adding a product.
type ProductsPostResponse struct {
	ProductId string `json:"productId" example:"prod123"`
}

// DeleteProductResponse godoc
// @Description Response returned after successful deletion of the product.
type DeleteProductResponse struct {
	Message string `json:"message" example:"product deleted successfully"`
}
