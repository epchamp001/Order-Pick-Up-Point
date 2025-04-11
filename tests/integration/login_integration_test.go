//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"order-pick-up-point/internal/models/dto"
)

func (s *TestSuite) TestLoginSuccess() {
	s.loadFixtures()

	payload := dto.LoginPostRequest{
		Email:    "login_success@example.com",
		Password: "strongpassword123",
	}
	tokenResp, status, err := s.loginUser(payload)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, status)
	s.Require().NotEmpty(tokenResp.Token)
}

// Логин с несуществующей почтой
func (s *TestSuite) TestLoginNonExistingEmail() {
	s.loadFixtures()
	payload := dto.LoginPostRequest{
		Email:    "nonexistent@example.com",
		Password: "strongpassword123",
	}
	_, status, err := s.loginUser(payload)
	s.Require().Error(err)
	s.Require().NotEqual(http.StatusOK, status)
}

// Логин с правильной почтой, но неверным паролем
func (s *TestSuite) TestLoginWrongPassword() {
	s.loadFixtures()
	payload := dto.LoginPostRequest{
		Email:    "login_success@example.com",
		Password: "wrongpassword",
	}
	_, status, err := s.loginUser(payload)
	s.Require().Error(err)
	s.Require().NotEqual(http.StatusOK, status)
}

// Некорректный JSON
func (s *TestSuite) TestLoginInvalidCredentials() {
	client := s.server.Client()
	url := fmt.Sprintf("%s/login", s.server.URL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(`{invalid json}`)))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().NotEqual(http.StatusOK, resp.StatusCode)

	var errResp dto.Error
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	s.Require().NoError(err)
	s.Require().NotEmpty(errResp.Message)
}
