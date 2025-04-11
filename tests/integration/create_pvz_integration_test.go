//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"order-pick-up-point/internal/models/dto"
)

func (s *TestSuite) TestCreatePvzSuccessForAllowedCities() {
	moderatorToken := s.getToken("moderator")
	allowedCities := []string{"Moscow", "Saint Petersburg", "Kazan"}

	for _, city := range allowedCities {
		pvzResp, status, err := s.createPvz(city, moderatorToken)
		s.Require().NoError(err, "city: %s", city)
		s.Require().Equal(http.StatusCreated, status, "city: %s", city)
		s.Require().NotEmpty(pvzResp.PvzId, "city: %s", city)
	}
}

// Клиент и сотрудник пытаются создать ПВЗ
func (s *TestSuite) TestCreatePvzNotAllowedForNonModerator() {
	clientToken := s.getToken("client")
	_, status, err := s.createPvz("Moscow", clientToken)
	s.Require().Error(err)
	s.Require().NotEqual(http.StatusCreated, status)

	employeeToken := s.getToken("employee")
	_, status, err = s.createPvz("Moscow", employeeToken)
	s.Require().Error(err)
	s.Require().NotEqual(http.StatusCreated, status)
}

// Создание ПВЗ в недопустимом городе
func (s *TestSuite) TestCreatePvzInvalidCity() {
	moderatorToken := s.getToken("moderator")
	_, status, err := s.createPvz("Новосибирск", moderatorToken)
	s.Require().Error(err)
	s.Require().NotEqual(http.StatusCreated, status)
}

// Отсутствие обязательных полей
func (s *TestSuite) TestCreatePvzMissingRequiredFields() {
	moderatorToken := s.getToken("moderator")
	// Пустой city
	client := s.server.Client()
	url := fmt.Sprintf("%s/pvz", s.server.URL)

	payload := dto.CreatePvzPostRequest{
		City: "",
	}
	body, err := json.Marshal(payload)
	s.Require().NoError(err)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", moderatorToken))

	resp, err := client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().NotEqual(http.StatusCreated, resp.StatusCode)

	var errResp dto.Error
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	s.Require().NoError(err)
	s.Require().NotEmpty(errResp.Message)
}
