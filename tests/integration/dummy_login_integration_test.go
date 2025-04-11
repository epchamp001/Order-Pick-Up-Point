//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"order-pick-up-point/internal/models/dto"
)

func (s *TestSuite) TestDummyLoginClient() {
	client := s.server.Client()
	url := fmt.Sprintf("%s/dummyLogin", s.server.URL)

	reqPayload := dto.DummyLoginPostRequest{
		Role: "client",
	}
	body, err := json.Marshal(reqPayload)
	s.Require().NoError(err)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var tokenResp dto.TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	resp.Body.Close()
	s.Require().NoError(err)
	s.Require().NotEmpty(tokenResp.Token)
}

func (s *TestSuite) TestDummyLoginModerator() {
	client := s.server.Client()
	url := fmt.Sprintf("%s/dummyLogin", s.server.URL)

	reqPayload := dto.DummyLoginPostRequest{
		Role: "moderator",
	}
	body, err := json.Marshal(reqPayload)
	s.Require().NoError(err)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var tokenResp dto.TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	resp.Body.Close()
	s.Require().NoError(err)
	s.Require().NotEmpty(tokenResp.Token)
}

func (s *TestSuite) TestDummyLoginEmployee() {
	client := s.server.Client()
	url := fmt.Sprintf("%s/dummyLogin", s.server.URL)

	reqPayload := dto.DummyLoginPostRequest{
		Role: "employee",
	}
	body, err := json.Marshal(reqPayload)
	s.Require().NoError(err)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var tokenResp dto.TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	resp.Body.Close()
	s.Require().NoError(err)
	s.Require().NotEmpty(tokenResp.Token)
}

func (s *TestSuite) TestDummyLoginInvalidRole() {
	client := s.server.Client()
	url := fmt.Sprintf("%s/dummyLogin", s.server.URL)

	reqPayload := dto.DummyLoginPostRequest{
		Role: "invalid_role",
	}
	body, err := json.Marshal(reqPayload)
	s.Require().NoError(err)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	s.Require().NoError(err)
	s.Require().NotEqual(http.StatusOK, resp.StatusCode)

	var errorResp dto.Error
	err = json.NewDecoder(resp.Body).Decode(&errorResp)
	resp.Body.Close()
	s.Require().NoError(err)
	s.Require().NotEmpty(errorResp.Message)
}
