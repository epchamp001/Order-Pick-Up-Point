//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"order-pick-up-point/internal/models/dto"
	"time"
)

func (s *TestSuite) TestGetPvzList_WithPagination_Success() {
	token := s.getToken("moderator")

	// создаём 3 ПВЗ
	for _, city := range []string{"Moscow", "Kazan", "Saint Petersburg"} {
		_, _, err := s.createPvz(city, token)
		s.Require().NoError(err)
	}

	// Запрос с limit=2 и page=1
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/pvz?page=1&limit=2", s.server.URL), nil)
	s.Require().NoError(err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := s.server.Client().Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var response []dto.PvzGet200ResponseInner
	err = json.NewDecoder(resp.Body).Decode(&response)
	s.Require().NoError(err)
	s.Require().Len(response, 2)
}

func (s *TestSuite) TestGetPvzList_WithReceptionDateFilter_Success() {
	token := s.getToken("moderator")

	// создаём ПВЗ
	pvzResp, _, err := s.createPvz("Moscow", token)
	s.Require().NoError(err)

	empToken := s.getToken("employee")

	// создаём приёмку со старой датой
	oldDate := time.Date(2023, 4, 1, 10, 0, 0, 0, time.UTC)
	_, _, err = s.createReception(pvzResp.PvzId, empToken, oldDate)
	s.Require().NoError(err)

	_, _, err = s.closeReception(pvzResp.PvzId, empToken)
	s.Require().NoError(err)

	// создаём приёмку с сегодняшней датой
	_, _, err = s.createReception(pvzResp.PvzId, empToken, time.Now())
	s.Require().NoError(err)

	_, _, err = s.closeReception(pvzResp.PvzId, empToken)
	s.Require().NoError(err)

	// задаём фильтр от 2023-04-01 до 2023-04-01
	start := time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	end := time.Date(2023, 4, 1, 23, 59, 59, 0, time.UTC).Format(time.RFC3339)

	req, err := http.NewRequest("GET", fmt.Sprintf(
		"%s/pvz?page=1&limit=10&startDate=%s&endDate=%s",
		s.server.URL, start, end,
	), nil)
	s.Require().NoError(err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := s.server.Client().Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var result []dto.PvzGet200ResponseInner
	err = json.NewDecoder(resp.Body).Decode(&result)
	s.Require().NoError(err)
	s.Require().NotEmpty(result)
	s.Require().Len(result[0].Receptions, 1)
}

func (s *TestSuite) TestGetPvzList_Unauthorized_Fail() {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/pvz", s.server.URL), nil)
	s.Require().NoError(err)

	resp, err := s.server.Client().Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (s *TestSuite) TestGetPvzList_InvalidPaginationParams_Fail() {
	token := s.getToken("moderator")

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/pvz?page=abc&limit=-1", s.server.URL), nil)
	s.Require().NoError(err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := s.server.Client().Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().NotEqual(http.StatusOK, resp.StatusCode)
}
