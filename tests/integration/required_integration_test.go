//go:build integration

package integration

import (
	"net/http"
	"time"
)

func (s *TestSuite) TestIntegrationFlow() {
	// Получаем токены для модератора и сотрудника через getToken
	modToken := s.getToken("moderator")
	empToken := s.getToken("employee")

	// 1. Создание ПВЗ модератором
	pvzResp, status, err := s.createPvz("Moscow", modToken)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusCreated, status)
	s.Require().NotEmpty(pvzResp.PvzId)

	// 2. Создание приёмки для созданного ПВЗ сотрудником
	recResp, status, err := s.createReception(pvzResp.PvzId, empToken, time.Now())
	s.Require().NoError(err)
	s.Require().Equal(http.StatusCreated, status)
	s.Require().NotEmpty(recResp.ReceptionId)

	// 3. Добавление 50 товаров с типом "electronics" сотрудником
	for i := 0; i < 50; i++ {
		prodResp, status, err := s.addProduct(pvzResp.PvzId, empToken, "electronics")
		s.Require().NoError(err, "iteration %d", i)
		s.Require().Equal(http.StatusCreated, status, "iteration %d", i)
		s.Require().NotEmpty(prodResp.ProductId, "iteration %d", i)
	}

	// 4. Закрытие приёмки сотрудником
	closeResp, status, err := s.closeReception(pvzResp.PvzId, empToken)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, status)
	s.Require().NotEmpty(closeResp.ReceptionId)
}
