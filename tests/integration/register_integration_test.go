//go:build integration

package integration

import (
	"net/http"
	"order-pick-up-point/internal/models/dto"
)

func (s *TestSuite) TestRegisterClientSuccess() {
	payload := dto.RegisterPostRequest{
		Email:    "client_new@example.com",
		Password: "strongpassword123",
		Role:     "client",
	}
	regResp, status, err := s.registerUser(payload)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusCreated, status)
	s.Require().NotEmpty(regResp.UserID)
}

func (s *TestSuite) TestRegisterModeratorSuccess() {
	payload := dto.RegisterPostRequest{
		Email:    "moderator_new@example.com",
		Password: "strongpassword123",
		Role:     "moderator",
	}
	regResp, status, err := s.registerUser(payload)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusCreated, status)
	s.Require().NotEmpty(regResp.UserID)
}

func (s *TestSuite) TestRegisterEmployeeSuccess() {
	payload := dto.RegisterPostRequest{
		Email:    "employee_new@example.com",
		Password: "strongpassword123",
		Role:     "employee",
	}
	regResp, status, err := s.registerUser(payload)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusCreated, status)
	s.Require().NotEmpty(regResp.UserID)
}

// Попытка регистрации с существующей почтой
func (s *TestSuite) TestRegisterExistingEmail() {
	s.loadFixtures()

	payload := dto.RegisterPostRequest{
		Email:    "duplicate@example.com",
		Password: "strongpassword123",
		Role:     "client",
	}
	_, status, err := s.registerUser(payload)
	s.Require().Error(err)
	s.Require().NotEqual(http.StatusCreated, status)
}

// Регистрация с некорректным email
func (s *TestSuite) TestRegisterInvalidEmail() {
	payload := dto.RegisterPostRequest{
		Email:    "user@",
		Password: "strongpassword123",
		Role:     "client",
	}
	_, status, err := s.registerUser(payload)
	s.Require().Error(err)
	s.Require().NotEqual(http.StatusCreated, status)
}

// Регистрация с коротким/пустым паролем
func (s *TestSuite) TestRegisterWeakPassword() {
	payloadEmpty := dto.RegisterPostRequest{
		Email:    "user_empty@example.com",
		Password: "",
		Role:     "client",
	}
	_, status, err := s.registerUser(payloadEmpty)
	s.Require().Error(err)
	s.Require().NotEqual(http.StatusCreated, status)

	payloadShort := dto.RegisterPostRequest{
		Email:    "user_short@example.com",
		Password: "123",
		Role:     "client",
	}
	_, status, err = s.registerUser(payloadShort)
	s.Require().Error(err)
	s.Require().NotEqual(http.StatusCreated, status)
}
