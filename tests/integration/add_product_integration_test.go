//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"order-pick-up-point/internal/models/dto"
	"time"
)

func (s *TestSuite) TestAddProductSuccess() {
	modToken := s.getToken("moderator")
	pvzResp, status, err := s.createPvz("Moscow", modToken)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusCreated, status)
	s.Require().NotEmpty(pvzResp.PvzId)

	empToken := s.getToken("employee")
	recResp, status, err := s.createReception(pvzResp.PvzId, empToken, time.Now())
	s.Require().NoError(err)
	s.Require().Equal(http.StatusCreated, status)
	s.Require().NotEmpty(recResp.ReceptionId)

	// Добавляем допустимый товар
	prodResp, status, err := s.addProduct(pvzResp.PvzId, empToken, "electronics")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusCreated, status)
	s.Require().NotEmpty(prodResp.ProductId)
}

// Добавляем товар без активной приёмки
func (s *TestSuite) TestAddProductWithoutActiveReception() {
	modToken := s.getToken("moderator")
	pvzResp, status, err := s.createPvz("Moscow", modToken)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusCreated, status)
	s.Require().NotEmpty(pvzResp.PvzId)

	empToken := s.getToken("employee")
	_, status, err = s.addProduct(pvzResp.PvzId, empToken, "electronics")
	s.Require().Error(err)
	s.Require().NotEqual(http.StatusCreated, status)
}

// Отсутствие заголовка авторизации
func (s *TestSuite) TestAddProductUnauthorized() {
	client := s.server.Client()
	url := fmt.Sprintf("%s/products", s.server.URL)

	payload := dto.ProductsPostRequest{
		PvzId: "some-pvz-id",
		Type:  "electronics",
	}
	body, err := json.Marshal(payload)
	s.Require().NoError(err)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	// Не устанавливаем заголовок Authorization

	resp, err := client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().NotEqual(http.StatusCreated, resp.StatusCode)
}

// Отсутствует обязательное поле данных товара
func (s *TestSuite) TestAddProductInvalidData() {
	modToken := s.getToken("moderator")
	pvzResp, status, err := s.createPvz("Moscow", modToken)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusCreated, status)
	s.Require().NotEmpty(pvzResp.PvzId)

	empToken := s.getToken("employee")
	recResp, status, err := s.createReception(pvzResp.PvzId, empToken, time.Now())
	s.Require().NoError(err)
	s.Require().Equal(http.StatusCreated, status)
	s.Require().NotEmpty(recResp.ReceptionId)

	// Некорректный payload — отсутствует поле "type"
	invalidPayload := map[string]interface{}{
		"pvzId": pvzResp.PvzId,
	}
	body, err := json.Marshal(invalidPayload)
	s.Require().NoError(err)

	client := s.server.Client()
	url := fmt.Sprintf("%s/products", s.server.URL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", empToken))

	resp, err := client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().NotEqual(http.StatusCreated, resp.StatusCode)
}
