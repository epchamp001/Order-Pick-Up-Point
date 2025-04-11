//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"order-pick-up-point/internal/models/dto"
	"time"
)

func (s *TestSuite) TestReceptionCreationByEmployeeSuccess() {
	modToken := s.getToken("moderator")
	pvzResp, status, err := s.createPvz("Moscow", modToken)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusCreated, status)
	s.Require().NotEmpty(pvzResp.PvzId)

	// Сотрудник создает приёмку для созданного ПВЗ
	empToken := s.getToken("employee")
	recResp, status, err := s.createReception(pvzResp.PvzId, empToken, time.Now())
	s.Require().NoError(err)
	s.Require().Equal(http.StatusCreated, status)
	s.Require().NotEmpty(recResp.ReceptionId)
}

// Попытка модератором и клиентом создавать приёмку
func (s *TestSuite) TestReceptionCreationNotAllowedForNonEmployee() {
	modToken := s.getToken("moderator")
	pvzResp, status, err := s.createPvz("Moscow", modToken)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusCreated, status)
	s.Require().NotEmpty(pvzResp.PvzId)

	modToken2 := s.getToken("moderator")
	_, status, err = s.createReception(pvzResp.PvzId, modToken2, time.Now())
	s.Require().Error(err)
	s.Require().NotEqual(http.StatusCreated, status)

	clientToken := s.getToken("client")
	_, status, err = s.createReception(pvzResp.PvzId, clientToken, time.Now())
	s.Require().Error(err)
	s.Require().NotEqual(http.StatusCreated, status)
}

// Попытка создать вторую приемку, при не закрытой первой
func (s *TestSuite) TestReceptionCreationWhenOpenReceptionExists() {
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

	// Попытка создания второй приёмки без закрытия первой
	_, status, err = s.createReception(pvzResp.PvzId, empToken, time.Now())
	s.Require().Error(err)
	s.Require().NotEqual(http.StatusCreated, status)
}

// Некорректный запрос, отсутствует pvzId
func (s *TestSuite) TestReceptionCreationMissingPvzID() {
	empToken := s.getToken("employee")
	client := s.server.Client()
	url := fmt.Sprintf("%s/receptions", s.server.URL)

	payload := dto.CreateReceptionRequest{
		PvzId:    "",
		DateTime: time.Now(),
	}
	body, err := json.Marshal(payload)
	s.Require().NoError(err)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", empToken))

	resp, err := client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().NotEqual(http.StatusCreated, resp.StatusCode)

	var errResp dto.Error
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	s.Require().NoError(err)
	s.Require().NotEmpty(errResp.Message)
}

// Создание приемки для несуществующего ПВЗ
func (s *TestSuite) TestReceptionCreationNonExistentPvz() {
	empToken := s.getToken("employee")
	// Генерируем валидный UUID, которого точно нет в базе
	nonExistentPvz := uuid.New().String()

	client := s.server.Client()
	url := fmt.Sprintf("%s/receptions", s.server.URL)

	payload := dto.CreateReceptionRequest{
		PvzId:    nonExistentPvz,
		DateTime: time.Now(),
	}
	body, err := json.Marshal(payload)
	s.Require().NoError(err)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", empToken))

	resp, err := client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().NotEqual(http.StatusCreated, resp.StatusCode)

	var errResp dto.Error
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	s.Require().NoError(err)
	s.Require().NotEmpty(errResp.Message)
}

// Создание приёмки с невалидным pvz_id
func (s *TestSuite) TestReceptionCreationInvalidPvzID() {
	empToken := s.getToken("employee")
	client := s.server.Client()
	url := fmt.Sprintf("%s/receptions", s.server.URL)

	payload := dto.CreateReceptionRequest{
		PvzId:    "invalid_uuid",
		DateTime: time.Now(),
	}
	body, err := json.Marshal(payload)
	s.Require().NoError(err)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", empToken))

	resp, err := client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().NotEqual(http.StatusCreated, resp.StatusCode)

	var errResp dto.Error
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	s.Require().NoError(err)
	s.Require().NotEmpty(errResp.Message)
}
