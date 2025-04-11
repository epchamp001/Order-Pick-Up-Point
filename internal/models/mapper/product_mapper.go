package mapper

import (
	"order-pick-up-point/internal/models/dto"
	"order-pick-up-point/internal/models/entity"
)

// ProductEntityToDTO преобразует сущность Product в DTO
func ProductEntityToDTO(p entity.Product) dto.ProductDTO {
	return dto.ProductDTO{
		Id:          p.ID,
		DateTime:    p.DateTime,
		Type:        p.Type,
		ReceptionId: p.ReceptionID,
	}
}

// ProductDTOToEntity преобразует DTO ProductDTO в сущность Product
func ProductDTOToEntity(p dto.ProductDTO) entity.Product {
	return entity.Product{
		ID:          p.Id,
		DateTime:    p.DateTime,
		Type:        p.Type,
		ReceptionID: p.ReceptionId,
	}
}
