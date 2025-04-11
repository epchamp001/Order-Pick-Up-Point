package mapper

import (
	"order-pick-up-point/internal/models/dto"
	"order-pick-up-point/internal/models/entity"
)

// ReceptionEntityToDTO преобразует сущность Reception в DTO.
func ReceptionEntityToDTO(r entity.Reception) dto.ReceptionDTO {
	return dto.ReceptionDTO{
		Id:       r.ID,
		DateTime: r.DateTime,
		PvzId:    r.PvzID,
		Status:   r.Status,
	}
}

// ReceptionDTOToEntity преобразует DTO ReceptionDTO в сущность Reception.
func ReceptionDTOToEntity(r dto.ReceptionDTO) entity.Reception {
	return entity.Reception{
		ID:       r.Id,
		DateTime: r.DateTime,
		PvzID:    r.PvzId,
		Status:   r.Status,
	}
}
