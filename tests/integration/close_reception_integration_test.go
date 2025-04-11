//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"time"
)

func (s *TestSuite) TestCloseReception_WithProducts_Success() {
	modToken := s.getToken("moderator")
	pvzResp, _, err := s.createPvz("Moscow", modToken)
	s.Require().NoError(err)

	empToken := s.getToken("employee")
	_, _, err = s.createReception(pvzResp.PvzId, empToken, time.Now())
	s.Require().NoError(err)

	// Добавим 1 товар
	_, status, err := s.addProduct(pvzResp.PvzId, empToken, "electronics")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusCreated, status)

	// Закрываем приёмку
	closeURL := fmt.Sprintf("%s/pvz/%s/close_last_reception", s.server.URL, pvzResp.PvzId)
	req, err := http.NewRequest("POST", closeURL, nil)
	s.Require().NoError(err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", empToken))

	resp, err := s.server.Client().Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)
}

func (s *TestSuite) TestCloseReception_Twice_Fail() {
	modToken := s.getToken("moderator")
	pvzResp, _, err := s.createPvz("Saint Petersburg", modToken)
	s.Require().NoError(err)

	empToken := s.getToken("employee")
	_, _, err = s.createReception(pvzResp.PvzId, empToken, time.Now())
	s.Require().NoError(err)

	// Добавим товар
	_, status, err := s.addProduct(pvzResp.PvzId, empToken, "clothes")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusCreated, status)

	// Первый раз закрываем — успех
	closeURL := fmt.Sprintf("%s/pvz/%s/close_last_reception", s.server.URL, pvzResp.PvzId)
	req, err := http.NewRequest("POST", closeURL, nil)
	s.Require().NoError(err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", empToken))
	resp, err := s.server.Client().Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	// Второй раз — должна быть ошибка
	req2, err := http.NewRequest("POST", closeURL, nil)
	s.Require().NoError(err)
	req2.Header.Set("Authorization", fmt.Sprintf("Bearer %s", empToken))
	resp2, err := s.server.Client().Do(req2)
	s.Require().NoError(err)
	defer resp2.Body.Close()
	s.Require().NotEqual(http.StatusOK, resp2.StatusCode)
}
