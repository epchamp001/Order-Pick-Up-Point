package db

type Repository interface {
	UserRepository
	PvzRepository
	ReceptionRepository
	ProductRepository
}

type postgresRepository struct {
	UserRepository
	PvzRepository
	ReceptionRepository
	ProductRepository
}

func NewRepository(
	userRepo UserRepository,
	pvzRepo PvzRepository,
	receptionRepo ReceptionRepository,
	productRepo ProductRepository,
) Repository {
	return &postgresRepository{
		UserRepository:      userRepo,
		PvzRepository:       pvzRepo,
		ReceptionRepository: receptionRepo,
		ProductRepository:   productRepo,
	}
}
