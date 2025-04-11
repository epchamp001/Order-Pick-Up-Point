package errs

const (
	// Общие
	ErrInvalidRequestCode = "INVALID_REQUEST"
	ErrNotFoundCode       = "NOT_FOUND"
	ErrUnauthorizedCode   = "UNAUTHORIZED"
	ErrInternalCode       = "INTERNAL_ERROR"

	// DummyLogin
	ErrInvalidRoleCode = "INVALID_ROLE" // неизвестная роль в /dummyLogin

	// Регистрация
	ErrUserAlreadyExists     = "USER_ALREADY_EXISTS"     // пользователь с таким email уже зарегистрирован
	ErrInvalidEmail          = "INVALID_EMAIL"           // email не соответствует формату
	ErrWeakPassword          = "WEAK_PASSWORD"           // пароль не удовлетворяет требованиям
	ErrPasswordHashingFailed = "PASSWORD_HASHING_FAILED" // ошибка при хэшировании пароля

	// Авторизация
	ErrInvalidCredentials = "INVALID_CREDENTIALS" // неверный email или пароль

	// Заведение ПВЗ
	ErrInvalidCity     = "INVALID_CITY"      // город не входит в допустимый список
	ErrForbiddenForPvz = "FORBIDDEN_FOR_PVZ" // пользователь без роли moderator пытается создать ПВЗ

	// Приёмка товаров
	ErrOpenReceptionExists = "OPEN_RECEPTION_EXISTS" // уже существует незакрытая приёмка для данного ПВЗ
	ErrNoOpenReception     = "NO_OPEN_RECEPTION"     // отсутствует незакрытая приёмка, к которой можно привязать товары

	// Товары
	ErrInvalidProductType = "INVALID_PRODUCT_TYPE"  // тип товара не является допустимым
	ErrNoProductsToDelete = "NO_PRODUCTS_TO_DELETE" // нет товаров для удаления в текущей незакрытой приёмке

	// Закрытие приёмки
	ErrReceptionAlreadyClosed = "RECEPTION_ALREADY_CLOSED" // приёмка уже закрыта
	ErrReceptionNotFound      = "RECEPTION_NOT_FOUND"      // приёмка не найдена для данного ПВЗ
)
