package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"net/url"
	"order-pick-up-point/internal/models/dto"
	"order-pick-up-point/internal/models/entity"
	"order-pick-up-point/internal/models/mapper"
	mockPvzServ "order-pick-up-point/internal/service/http/mock"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestCheckRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		contextRole     interface{} // Значение, которое устанавливается в c.Set("role", ...)
		allowedRoles    []string
		expected        bool
		expectedStatus  int    // ожидаемый HTTP статус при ошибке
		expectedMessage string // ожидаемое сообщение в JSON-ответе (при ошибке)
	}{
		{
			name:            "missing role",
			contextRole:     nil,
			allowedRoles:    []string{"client"},
			expected:        false,
			expectedStatus:  http.StatusUnauthorized,
			expectedMessage: "missing role in token",
		},
		{
			name:            "invalid role type",
			contextRole:     123, // не строка
			allowedRoles:    []string{"client"},
			expected:        false,
			expectedStatus:  http.StatusForbidden,
			expectedMessage: "invalid role type",
		},
		{
			name:            "role not allowed",
			contextRole:     "employee",
			allowedRoles:    []string{"client"},
			expected:        false,
			expectedStatus:  http.StatusForbidden,
			expectedMessage: "access denied, allowed roles: [client]",
		},
		{
			name:            "role allowed",
			contextRole:     "Client", // проверка нечувствительность к регистру
			allowedRoles:    []string{"client", "moderator"},
			expected:        true,
			expectedStatus:  0, // ответ не формируется
			expectedMessage: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rr)

			// Если значение не равно nil, устанавливаем его в контексте
			if tc.contextRole != nil {
				c.Set("role", tc.contextRole)
			}

			// Вызываем функцию CheckRole с заданными allowedRoles
			result := CheckRole(c, tc.allowedRoles...)

			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
			// Если функция возвращает false (то есть произошла ошибка), проверяем, что в ответе записан JSON с нужным статусом и сообщением
			if !result {
				if rr.Code != tc.expectedStatus {
					t.Errorf("expected status %d, got %d", tc.expectedStatus, rr.Code)
				}
				body := rr.Body.String()
				if !strings.Contains(body, tc.expectedMessage) {
					t.Errorf("expected response to contain %q, got %q", tc.expectedMessage, body)
				}
			}
		})
	}
}

func TestPvzController_CreatePvz(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		requestBody  string
		unauthorized bool
		// если simulateSvcError==true, то p.pvzSvc.CreatePvz вернет ошибку
		simulateSvcError   bool
		svcErr             error
		expectedPvzID      string // ожидаемый ID, если успех
		expectedStatusCode int
		expectedRespSubstr string // ожидаемая подстрока в JSON-ответе
	}{
		{
			name:               "unauthorized - missing role",
			requestBody:        `{"city": "Moscow"}`,
			unauthorized:       true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedRespSubstr: "missing role in token",
		},
		{
			name:               "invalid request body",
			requestBody:        "invalid json",
			unauthorized:       false,
			expectedStatusCode: http.StatusBadRequest,
			expectedRespSubstr: "invalid request body",
		},
		{
			name:               "service error",
			requestBody:        `{"city": "Moscow"}`,
			unauthorized:       false,
			simulateSvcError:   true,
			svcErr:             errors.New("create pvz error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedRespSubstr: "create pvz error",
		},
		{
			name:               "success",
			requestBody:        `{"city": "Moscow"}`,
			unauthorized:       false,
			simulateSvcError:   false,
			expectedPvzID:      "pvz123",
			expectedStatusCode: http.StatusCreated,
			expectedRespSubstr: `"id":"pvz123"`,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest("POST", "/pvz", bytes.NewBufferString(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(rr)
			c.Request = req

			// Если тестовый сценарий авторизован, устанавливаем роль "moderator" в контексте
			// Если unauthorized == true, роль не устанавливается
			if !tc.unauthorized {
				c.Set("role", "moderator")
			}

			mockSvc := mockPvzServ.NewPvzService(t)

			// Если unauthorized, контроллер не должен вызывать метод CreatePvz
			if !tc.unauthorized {
				// Если тело запроса корректное (начинается с "{"),
				// и ожидается вызов сервиса
				if strings.HasPrefix(tc.requestBody, "{") {
					if tc.simulateSvcError {
						mockSvc.
							On("CreatePvz", mock.Anything, "Moscow").
							Return("", tc.svcErr).
							Once()
					} else {
						mockSvc.
							On("CreatePvz", mock.Anything, "Moscow").
							Return(tc.expectedPvzID, nil).
							Once()
					}
				}
			}

			ctrl := NewPvzController(mockSvc)
			ctrl.CreatePvz(c)

			// Если unauthorized, ожидаем, что ответ уже сформирован и тело содержит соответствующее сообщение
			if tc.unauthorized {
				if rr.Code != http.StatusUnauthorized && rr.Code != http.StatusForbidden {
					t.Errorf("expected unauthorized status, got %d", rr.Code)
				}
				if !strings.Contains(rr.Body.String(), "missing role in token") &&
					!strings.Contains(rr.Body.String(), "access denied") {
					t.Errorf("expected response containing unauthorized message, got %s", rr.Body.String())
				}
				return
			}

			// Проверяем, если тело запроса невалидное
			if tc.requestBody == "invalid json" {
				if rr.Code != http.StatusBadRequest {
					t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
				}
				if !strings.Contains(rr.Body.String(), tc.expectedRespSubstr) {
					t.Errorf("expected response containing %q, got %q", tc.expectedRespSubstr, rr.Body.String())
				}
				return
			}

			// Если сервис возвращает ошибку
			if tc.simulateSvcError {
				if rr.Code != http.StatusInternalServerError {
					t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
				}
				if !strings.Contains(rr.Body.String(), tc.expectedRespSubstr) {
					t.Errorf("expected response containing %q, got %q", tc.expectedRespSubstr, rr.Body.String())
				}
			} else {
				// Успешный сценарий
				if rr.Code != http.StatusCreated {
					t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
				}
				if !strings.Contains(rr.Body.String(), tc.expectedRespSubstr) {
					t.Errorf("expected response containing %q, got %q", tc.expectedRespSubstr, rr.Body.String())
				}
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestPvzController_GetPvzsInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Now()
	dummyPvzInfos := []entity.PvzInfo{
		{
			Pvz: entity.Pvz{
				ID:               "pvz1",
				City:             "Moscow",
				RegistrationDate: now,
			},
			Receptions: []entity.ReceptionInfo{},
		},
		{
			Pvz: entity.Pvz{
				ID:               "pvz2",
				City:             "Kazan",
				RegistrationDate: now.Add(-time.Hour),
			},
			Receptions: []entity.ReceptionInfo{},
		},
	}

	expectedResponse := make([]dto.PvzGet200ResponseInner, 0, len(dummyPvzInfos))
	for _, info := range dummyPvzInfos {
		expectedResponse = append(expectedResponse, mapper.PvzInfoEntityToResponse(info))
	}

	tests := []struct {
		name               string
		queryParams        map[string]string
		simulateError      bool
		svcErr             error
		expectedStatusCode int
		expectedRespSubstr string
	}{
		{
			name: "missing page parameter",
			queryParams: map[string]string{
				"limit": "10",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedRespSubstr: "missing page or limit parameter",
		},
		{
			name: "invalid page parameter",
			queryParams: map[string]string{
				"page":  "abc",
				"limit": "10",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedRespSubstr: "invalid page parameter",
		},
		{
			name: "invalid limit parameter",
			queryParams: map[string]string{
				"page":  "1",
				"limit": "xyz",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedRespSubstr: "invalid limit parameter",
		},
		{
			name: "invalid startDate format",
			queryParams: map[string]string{
				"page":      "1",
				"limit":     "10",
				"startDate": "2025-04-09", // не RFC3339
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedRespSubstr: "invalid startDate format",
		},
		{
			name: "invalid endDate format",
			queryParams: map[string]string{
				"page":    "1",
				"limit":   "10",
				"endDate": "2025/04/09 23:59:59", // неверный формат
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedRespSubstr: "invalid endDate format",
		},
		{
			name: "service error",
			queryParams: map[string]string{
				"page":      "1",
				"limit":     "10",
				"startDate": "2025-04-09T00:00:00Z",
				"endDate":   "2025-04-09T23:59:59Z",
			},
			simulateError:      true,
			svcErr:             fmt.Errorf("database failure"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedRespSubstr: "failed to get PVZ info",
		},
		{
			name: "success",
			queryParams: map[string]string{
				"page":      "1",
				"limit":     "10",
				"startDate": "2025-04-09T00:00:00Z",
				"endDate":   "2025-04-09T23:59:59Z",
			},
			simulateError:      false,
			svcErr:             nil,
			expectedStatusCode: http.StatusOK,
			expectedRespSubstr: "pvz1",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Формируем query-параметры через url.Values.
			values := url.Values{}
			for key, value := range tc.queryParams {
				values.Set(key, value)
			}
			queryStr := values.Encode()

			req := httptest.NewRequest("GET", "/pvz?"+queryStr, nil)
			rr := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(rr)
			c.Request = req

			c.Set("role", "moderator")

			mockSvc := mockPvzServ.NewPvzService(t)

			// Если параметры page и limit заданы, то извлекаем их
			if pageStr, ok1 := tc.queryParams["page"]; ok1 && pageStr != "" {
				if limitStr, ok2 := tc.queryParams["limit"]; ok2 && limitStr != "" && tc.expectedStatusCode != http.StatusBadRequest {
					page, _ := strconv.Atoi(pageStr)
					limit, _ := strconv.Atoi(limitStr)
					if tc.simulateError {
						mockSvc.
							On("GetPvzsInfo", mock.Anything, page, limit, mock.Anything, mock.Anything).
							Return(nil, tc.svcErr).
							Once()
					} else {
						mockSvc.
							On("GetPvzsInfo", mock.Anything, page, limit, mock.Anything, mock.Anything).
							Return(dummyPvzInfos, nil).
							Once()
					}
				}
			}

			ctrl := NewPvzController(mockSvc)
			ctrl.GetPvzsInfo(c)

			if rr.Code != tc.expectedStatusCode {
				t.Errorf("expected status code %d, got %d", tc.expectedStatusCode, rr.Code)
			}

			respBody := strings.TrimSpace(rr.Body.String())
			if !strings.Contains(respBody, tc.expectedRespSubstr) {
				t.Errorf("expected response containing %q, got %q", tc.expectedRespSubstr, respBody)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestPvzController_CreateReception(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receptionTime := time.Date(2025, time.April, 10, 12, 0, 0, 0, time.UTC)
	requestObj := dto.CreateReceptionRequest{
		PvzId:    "pvz123",
		DateTime: receptionTime,
	}
	reqBodyBytes, _ := json.Marshal(requestObj)

	tests := []struct {
		name        string
		requestBody string
		// Если unauthorized==true, роль не устанавливается
		unauthorized bool
		// Если simulateSvcError==true, то pvzSvc.CreateReception возвращает ошибку
		simulateSvcError    bool
		svcErr              error
		expectedReceptionID string
		expectedStatusCode  int
		expectedRespSubstr  string
	}{
		{
			name:               "unauthorized: missing role",
			requestBody:        `{"pvzId": "pvz123", "dateTime": "2025-04-10T12:00:00Z"}`,
			unauthorized:       true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedRespSubstr: "missing role",
		},
		{
			name:               "invalid request body",
			requestBody:        "invalid json",
			unauthorized:       false,
			expectedStatusCode: http.StatusBadRequest,
			expectedRespSubstr: "invalid request body",
		},
		{
			name:               "service error",
			requestBody:        `{"pvzId": "pvz123", "dateTime": "2025-04-10T12:00:00Z"}`,
			unauthorized:       false,
			simulateSvcError:   true,
			svcErr:             errors.New("create reception error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedRespSubstr: "create reception error",
		},
		{
			name:                "success",
			requestBody:         string(reqBodyBytes),
			unauthorized:        false,
			simulateSvcError:    false,
			expectedReceptionID: "rec789",
			expectedStatusCode:  http.StatusCreated,
			expectedRespSubstr:  `"receptionId":"rec789"`,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest("POST", "/receptions", bytes.NewBufferString(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(rr)
			c.Request = req

			// Если не unauthorized, устанавливаем роль "employee"
			if !tc.unauthorized {
				c.Set("role", "employee")
			}

			mockSvc := mockPvzServ.NewPvzService(t)
			if !tc.unauthorized && tc.requestBody != "invalid json" {
				if tc.simulateSvcError {
					mockSvc.
						On("CreateReception", mock.Anything, "pvz123", receptionTime).
						Return("", tc.svcErr).
						Once()
				} else {
					mockSvc.
						On("CreateReception", mock.Anything, "pvz123", receptionTime).
						Return(tc.expectedReceptionID, nil).
						Once()
				}
			}

			ctrl := NewPvzController(mockSvc)
			ctrl.CreateReception(c)

			// Если unauthorized, контроллер должен завершиться
			if tc.unauthorized {
				if rr.Code != http.StatusUnauthorized && rr.Code != http.StatusForbidden {
					t.Errorf("expected unauthorized status, got %d", rr.Code)
				}
				if !strings.Contains(rr.Body.String(), "missing role in token") &&
					!strings.Contains(rr.Body.String(), "access denied") {
					t.Errorf("expected unauthorized message, got %s", rr.Body.String())
				}
				return
			}

			// Если тело запроса невалидное
			if tc.requestBody == "invalid json" {
				if rr.Code != http.StatusBadRequest {
					t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
				}
				if !strings.Contains(rr.Body.String(), "invalid request body") {
					t.Errorf("expected response containing %q, got %q", "invalid request body", rr.Body.String())
				}
				return
			}

			// Если сервис возвращает ошибку
			if tc.simulateSvcError {
				if rr.Code != http.StatusInternalServerError {
					t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
				}
				if !strings.Contains(rr.Body.String(), tc.expectedRespSubstr) {
					t.Errorf("expected response containing %q, got %q", tc.expectedRespSubstr, rr.Body.String())
				}
			} else {
				if rr.Code != http.StatusCreated {
					t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
				}
				if !strings.Contains(rr.Body.String(), tc.expectedRespSubstr) {
					t.Errorf("expected response containing %q, got %q", tc.expectedRespSubstr, rr.Body.String())
				}
			}
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestPvzController_AddProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		requestBody        string
		unauthorized       bool   // если true, роль не устанавливается
		simulateSvcError   bool   // если true, p.pvzSvc.AddProduct возвращает ошибку
		svcErr             error  // ошибка, которую возвращает сервис
		expectedProductID  string // ожидаемый productID в успешном случае
		expectedStatusCode int
		expectedRespSubstr string
	}{
		{
			name:               "unauthorized: missing role",
			requestBody:        `{"pvzId": "pvz123", "type": "electronics"}`,
			unauthorized:       true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedRespSubstr: "missing role",
		},
		{
			name:               "invalid request body",
			requestBody:        "invalid json",
			unauthorized:       false,
			expectedStatusCode: http.StatusBadRequest,
			expectedRespSubstr: "invalid request body",
		},
		{
			name:               "service error",
			requestBody:        `{"pvzId": "pvz123", "type": "electronics"}`,
			unauthorized:       false,
			simulateSvcError:   true,
			svcErr:             errors.New("add product error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedRespSubstr: "add product error",
		},
		{
			name:               "success",
			requestBody:        `{"pvzId": "pvz123", "type": "electronics"}`,
			unauthorized:       false,
			simulateSvcError:   false,
			expectedProductID:  "prod789",
			expectedStatusCode: http.StatusCreated,
			expectedRespSubstr: `"productId":"prod789"`,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest("POST", "/products", bytes.NewBufferString(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(rr)
			c.Request = req

			// Если тестовый сценарий авторизован, устанавливаем роль "employee"
			if !tc.unauthorized {
				c.Set("role", "employee")
			}

			// Создаем мок-сервис для PVZ, используя сгенерированную функцию NewPvzService
			mockSvc := mockPvzServ.NewPvzService(t)
			// Если requestBody корректный (начинается с "{"), то ожидаем вызов метода
			if strings.HasPrefix(tc.requestBody, "{") && !tc.unauthorized {
				if tc.simulateSvcError {
					mockSvc.
						On("AddProduct", mock.Anything, "pvz123", "electronics").
						Return("", tc.svcErr).
						Once()
				} else {
					mockSvc.
						On("AddProduct", mock.Anything, "pvz123", "electronics").
						Return(tc.expectedProductID, nil).
						Once()
				}
			}

			ctrl := NewPvzController(mockSvc)
			ctrl.AddProduct(c)

			if rr.Code != tc.expectedStatusCode {
				t.Errorf("expected status code %d, got %d", tc.expectedStatusCode, rr.Code)
			}

			respBody := strings.TrimSpace(rr.Body.String())
			if !strings.Contains(respBody, tc.expectedRespSubstr) {
				t.Errorf("expected response containing %q, got %q", tc.expectedRespSubstr, respBody)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestPvzController_DeleteLastProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		pvzId              string
		unauthorized       bool  // если true – роль не устанавливается
		simulateSvcError   bool  // если true – сервис возвращает ошибку
		svcErr             error // ошибка, которую возвращает сервис
		expectedStatusCode int
		expectedRespSubstr string
	}{
		{
			name:               "unauthorized: missing role",
			pvzId:              "pvz123",
			unauthorized:       true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedRespSubstr: "missing role",
		},
		{
			name:               "missing pvzId parameter",
			pvzId:              "",
			unauthorized:       false,
			expectedStatusCode: http.StatusBadRequest,
			expectedRespSubstr: "pvzId parameter is required",
		},
		{
			name:               "service error",
			pvzId:              "pvz123",
			unauthorized:       false,
			simulateSvcError:   true,
			svcErr:             errors.New("delete error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedRespSubstr: "delete error",
		},
		{
			name:               "success",
			pvzId:              "pvz123",
			unauthorized:       false,
			simulateSvcError:   false,
			expectedStatusCode: http.StatusOK,
			expectedRespSubstr: "product deleted successfully",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest("POST", "/pvz/"+tc.pvzId+"/delete_last_product", nil)
			rr := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(rr)
			c.Request = req

			// Устанавливаем параметр пути "pvzId"
			c.Params = gin.Params{{
				Key:   "pvzId",
				Value: tc.pvzId,
			}}

			// Если сценарий авторизованный, устанавливаем роль "employee" в контексте
			if !tc.unauthorized {
				c.Set("role", "employee")
			}

			mockSvc := mockPvzServ.NewPvzService(t)
			// Если pvzId задан и сценарий не unauthorized,
			// то ожидаем вызов метода DeleteLastProduct
			if !tc.unauthorized && tc.pvzId != "" {
				if tc.simulateSvcError {
					mockSvc.
						On("DeleteLastProduct", mock.Anything, tc.pvzId).
						Return(errors.New("delete error")).
						Once()
				} else {
					mockSvc.
						On("DeleteLastProduct", mock.Anything, tc.pvzId).
						Return(nil).
						Once()
				}
			}

			ctrl := NewPvzController(mockSvc)
			ctrl.DeleteLastProduct(c)

			if rr.Code != tc.expectedStatusCode {
				t.Errorf("expected status code %d, got %d", tc.expectedStatusCode, rr.Code)
			}

			respBody := strings.TrimSpace(rr.Body.String())
			if !strings.Contains(respBody, tc.expectedRespSubstr) {
				t.Errorf("expected response to contain %q, got %q", tc.expectedRespSubstr, respBody)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestPvzController_CloseReception(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		pathParam          string // значение, которое вернет c.Param("pvzId")
		unauthorized       bool   // если true, роль не устанавливается в контексте
		simulateSvcError   bool   // если true, p.pvzSvc.CloseReception возвращает ошибку
		svcErr             error  // ошибка, которую возвращает сервис
		expectedStatusCode int
		expectedRespSubstr string
	}{
		{
			name:               "unauthorized - missing role",
			pathParam:          "pvz123",
			unauthorized:       true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedRespSubstr: "missing role",
		},
		{
			name:               "missing pvzId parameter",
			pathParam:          "",
			unauthorized:       false,
			expectedStatusCode: http.StatusBadRequest,
			expectedRespSubstr: "pvzId parameter is required",
		},
		{
			name:               "service error",
			pathParam:          "pvz123",
			unauthorized:       false,
			simulateSvcError:   true,
			svcErr:             errors.New("close reception error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedRespSubstr: "close reception error",
		},
		{
			name:               "success",
			pathParam:          "pvz123",
			unauthorized:       false,
			simulateSvcError:   false,
			svcErr:             nil,
			expectedStatusCode: http.StatusOK,
			expectedRespSubstr: `"receptionId":"pvz123"`,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest("POST", "/pvz/"+tc.pathParam+"/close_last_reception", nil)
			rr := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(rr)
			c.Request = req

			// Устанавливаем параметр пути "pvzId"
			c.Params = gin.Params{
				{Key: "pvzId", Value: tc.pathParam},
			}

			// Если сценарий авторизованный, устанавливаем роль "employee"
			if !tc.unauthorized {
				c.Set("role", "employee")
			}

			mockSvc := mockPvzServ.NewPvzService(t)
			if !tc.unauthorized && tc.pathParam != "" {
				if tc.simulateSvcError {
					mockSvc.
						On("CloseReception", mock.Anything, tc.pathParam).
						Return("", tc.svcErr).
						Once()
				} else {
					// В успешном сценарии сервис возвращает ID закрытой приёмки
					mockSvc.
						On("CloseReception", mock.Anything, tc.pathParam).
						Return(tc.pathParam, nil).
						Once()
				}
			}

			ctrl := NewPvzController(mockSvc)
			ctrl.CloseReception(c)

			if rr.Code != tc.expectedStatusCode {
				t.Errorf("expected status code %d, got %d", tc.expectedStatusCode, rr.Code)
			}

			respBody := strings.TrimSpace(rr.Body.String())
			if !strings.Contains(respBody, tc.expectedRespSubstr) {
				t.Errorf("expected response containing %q, got %q", tc.expectedRespSubstr, respBody)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}
