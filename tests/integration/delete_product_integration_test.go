//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"time"
)

func (s *TestSuite) TestDeleteLastProduct_Success() {
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

	// Удалим последний товар
	deleteURL := fmt.Sprintf("%s/pvz/%s/delete_last_product", s.server.URL, pvzResp.PvzId)
	req, err := http.NewRequest("POST", deleteURL, nil)
	s.Require().NoError(err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", empToken))

	resp, err := s.server.Client().Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)
}

func (s *TestSuite) TestDeleteAllProducts_Sequentially() {
	modToken := s.getToken("moderator")
	pvzResp, _, err := s.createPvz("Moscow", modToken)
	s.Require().NoError(err)

	empToken := s.getToken("employee")
	_, _, err = s.createReception(pvzResp.PvzId, empToken, time.Now())
	s.Require().NoError(err)

	// Добавим 3 товара
	for i := 0; i < 3; i++ {
		_, status, err := s.addProduct(pvzResp.PvzId, empToken, "clothes")
		s.Require().NoError(err)
		s.Require().Equal(http.StatusCreated, status)
	}

	// Удалим по одному
	for i := 0; i < 3; i++ {
		deleteURL := fmt.Sprintf("%s/pvz/%s/delete_last_product", s.server.URL, pvzResp.PvzId)
		req, err := http.NewRequest("POST", deleteURL, nil)
		s.Require().NoError(err)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", empToken))

		resp, err := s.server.Client().Do(req)
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Require().Equal(http.StatusOK, resp.StatusCode)
	}
}

func (s *TestSuite) TestDeleteProductFromEmptyReception_Fail() {
	modToken := s.getToken("moderator")
	pvzResp, _, err := s.createPvz("Kazan", modToken)
	s.Require().NoError(err)

	empToken := s.getToken("employee")
	_, _, err = s.createReception(pvzResp.PvzId, empToken, time.Now())
	s.Require().NoError(err)

	// Удаляем, но ничего не добавляли
	deleteURL := fmt.Sprintf("%s/pvz/%s/delete_last_product", s.server.URL, pvzResp.PvzId)
	req, err := http.NewRequest("POST", deleteURL, nil)
	s.Require().NoError(err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", empToken))

	resp, err := s.server.Client().Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().NotEqual(http.StatusOK, resp.StatusCode)
}
