package mapper

import (
	"order-pick-up-point/internal/models/dto"
	"order-pick-up-point/internal/models/entity"
)

// PvzEntityToDTO преобразует сущность Pvz в DTO.
func PvzEntityToDTO(p entity.Pvz) dto.PvzDTO {
	return dto.PvzDTO{
		Id:               p.ID,
		RegistrationDate: p.RegistrationDate,
		City:             p.City,
	}
}

// PvzDTOToEntity преобразует DTO PvzDTO в сущность Pvz.
func PvzDTOToEntity(p dto.PvzDTO) entity.Pvz {
	return entity.Pvz{
		ID:               p.Id,
		RegistrationDate: p.RegistrationDate,
		City:             p.City,
	}
}

func PvzInfoEntityToResponse(info entity.PvzInfo) dto.PvzGet200ResponseInner {
	pvzDTO := PvzEntityToDTO(info.Pvz)

	receptions := make([]dto.PvzGet200ResponseInnerReceptionsInner, 0, len(info.Receptions))
	for _, recInfo := range info.Receptions {
		recDTO := ReceptionEntityToDTO(recInfo.Reception)
		products := make([]dto.ProductDTO, 0, len(recInfo.Products))
		for _, prod := range recInfo.Products {
			products = append(products, ProductEntityToDTO(prod))
		}

		receptions = append(receptions, dto.PvzGet200ResponseInnerReceptionsInner{
			Reception: &recDTO,
			Products:  products,
		})
	}

	return dto.PvzGet200ResponseInner{
		Pvz:        &pvzDTO,
		Receptions: receptions,
	}
}
