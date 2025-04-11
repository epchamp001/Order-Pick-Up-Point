package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	mockAuthServ "order-pick-up-point/internal/service/http/mock"
	"strings"
	"testing"
)

func TestAuthController_DummyLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		requestBody      string
		svcReturnToken   string
		svcReturnErr     error
		expectedStatus   int
		expectedResponse string
	}{
		{
			name:             "invalid request body",
			requestBody:      "invalid json",
			svcReturnToken:   "",
			svcReturnErr:     nil,
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"message":"invalid request body"}`,
		},
		{
			name:           "service returns error",
			requestBody:    `{"role": "client"}`,
			svcReturnToken: "",
			svcReturnErr:   errors.New("dummy login failed"),
			expectedStatus: http.StatusUnauthorized,
			// Контроллер возвращает текст ошибки через err.Error()
			expectedResponse: `{"message":"dummy login failed"}`,
		},
		{
			name:             "success",
			requestBody:      `{"role": "client"}`,
			svcReturnToken:   "jwt_token",
			svcReturnErr:     nil,
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"token":"jwt_token"}`,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/dummyLogin", strings.NewReader(tc.requestBody))
			c.Request.Header.Set("Content-Type", "application/json")

			mockSvc := mockAuthServ.NewAuthService(t)

			// Если тело запроса корректное, контроллер вызовет метод DummyLogin
			if strings.HasPrefix(tc.requestBody, "{") {
				// Из JSON ожидается поле role, здесь "client"
				mockSvc.
					On("DummyLogin", mock.Anything, "client").
					Return(tc.svcReturnToken, tc.svcReturnErr).
					Once()
			}

			ctrl := NewAuthController(mockSvc)
			ctrl.DummyLogin(c)

			if w.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, w.Code)
			}
			// Проверяем, что тело ответа содержит ожидаемую подстроку
			gotBody := strings.TrimSpace(w.Body.String())
			if !strings.Contains(gotBody, tc.expectedResponse) {
				t.Errorf("expected response containing %s, got %s", tc.expectedResponse, gotBody)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestAuthController_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name               string
		requestBody        string
		expectedErrMsg     string
		expectedStatusCode int
		// Если simulateSvcError == true, то мок authSvc.Register вернет ошибку, иначе ожидается успешная регистрация
		simulateSvcError bool
		svcErr           error
		expectedUserID   string
	}{
		{
			name:               "invalid request body",
			requestBody:        "invalid json",
			expectedErrMsg:     "invalid request body",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "invalid email",
			requestBody:        `{"email": "invalid", "password": "secret", "role": "client"}`,
			expectedErrMsg:     "invalid email format",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "invalid password",
			requestBody:        `{"email": "test@example.com", "password": "123", "role": "client"}`,
			expectedErrMsg:     "password must be at least 4 characters long",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "service registration error",
			requestBody:        `{"email": "test@example.com", "password": "secret", "role": "client"}`,
			expectedErrMsg:     "create error",
			expectedStatusCode: http.StatusInternalServerError,
			simulateSvcError:   true,
			svcErr:             errors.New("create error"),
		},
		{
			name:               "success",
			requestBody:        `{"email": "test@example.com", "password": "secret", "role": "client"}`,
			expectedStatusCode: http.StatusCreated,
			simulateSvcError:   false,
			expectedUserID:     "user123",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest("POST", "/register", bytes.NewBufferString(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(rr)
			c.Request = req

			mockAuthSvc := mockAuthServ.NewAuthService(t)

			// Если тело запроса корректное (начинается с "{"), то ожидаем вызов метода Register, если валидаторы пропускают запрос
			// Для сценариев invalid email/password, валидаторы должны сработать до вызова authSvc.Register.
			if strings.HasPrefix(tc.requestBody, "{") && tc.expectedStatusCode == http.StatusCreated || tc.simulateSvcError {
				mockAuthSvc.
					On("Register", mock.Anything, "test@example.com", "secret", "client").
					Return(func(ctx context.Context, email, password, role string) string {
						return tc.expectedUserID
					}, tc.svcErr).
					Once()
			}

			ctrl := NewAuthController(mockAuthSvc)
			ctrl.Register(c)

			if rr.Code != tc.expectedStatusCode {
				t.Errorf("expected status %d, got %d", tc.expectedStatusCode, rr.Code)
			}

			respBody := strings.TrimSpace(rr.Body.String())

			if tc.expectedErrMsg != "" {
				if !strings.Contains(respBody, tc.expectedErrMsg) {
					t.Errorf("expected response containing %q, got %q", tc.expectedErrMsg, respBody)
				}
			} else {
				// В успешном случае ожидаем JSON с ключом "user_id"
				expectedResponse := fmt.Sprintf(`{"user_id":"%s"}`, tc.expectedUserID)
				if !strings.Contains(respBody, expectedResponse) {
					t.Errorf("expected response containing %q, got %q", expectedResponse, respBody)
				}
			}

			mockAuthSvc.AssertExpectations(t)
		})
	}
}

func TestAuthController_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name               string
		requestBody        string
		expectedStatusCode int
		expectedErrMsg     string
		// Если simulateSvcError==true, то мок authSvc.Login возвращает ошибку, иначе возвращает токен
		simulateSvcError bool
		svcErr           error
		expectedToken    string
	}{
		{
			name:               "invalid request body",
			requestBody:        "not a json",
			expectedStatusCode: http.StatusBadRequest,
			expectedErrMsg:     "invalid request body",
		},
		{
			name:               "service login error",
			requestBody:        `{"email": "test@example.com", "password": "secret"}`,
			expectedStatusCode: http.StatusUnauthorized,
			expectedErrMsg:     "invalid credentials",
			simulateSvcError:   true,
			svcErr:             errors.New("invalid credentials"),
		},
		{
			name:               "success",
			requestBody:        `{"email": "test@example.com", "password": "secret"}`,
			expectedStatusCode: http.StatusOK,
			expectedToken:      "jwt_token_abc",
			simulateSvcError:   false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(rr)
			c.Request = req

			mockAuthSvc := mockAuthServ.NewAuthService(t)

			// Если тело запроса корректное, ожидаем вызов метода Login
			if strings.HasPrefix(tc.requestBody, "{") {
				mockAuthSvc.
					On("Login", mock.Anything, "test@example.com", "secret").
					Return(func(ctx context.Context, email, password string) string {
						return tc.expectedToken
					}, tc.svcErr).
					Once()
			}

			ctrl := NewAuthController(mockAuthSvc)
			ctrl.Login(c)

			if rr.Code != tc.expectedStatusCode {
				t.Errorf("expected status code %d, got %d", tc.expectedStatusCode, rr.Code)
			}

			respBody := strings.TrimSpace(rr.Body.String())
			if tc.expectedErrMsg != "" {
				if !strings.Contains(respBody, tc.expectedErrMsg) {
					t.Errorf("expected response containing %q, got %q", tc.expectedErrMsg, respBody)
				}
			} else {
				// Ожидаем JSON, содержащий поле "token"
				expectedStr := fmt.Sprintf(`"token":"%s"`, tc.expectedToken)
				if !strings.Contains(respBody, expectedStr) {
					t.Errorf("expected response containing %q, got %q", expectedStr, respBody)
				}
			}

			mockAuthSvc.AssertExpectations(t)
		})
	}
}
