//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"order-pick-up-point/internal/models/dto"
	"time"
)

// loginUser отправляет POST-запрос на /login с заданными учетными данными
func (s *TestSuite) loginUser(payload dto.LoginPostRequest) (*dto.TokenResponse, int, error) {
	client := s.server.Client()
	url := fmt.Sprintf("%s/login", s.server.URL)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var tokenResp dto.TokenResponse
		if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
			return nil, resp.StatusCode, err
		}
		return &tokenResp, resp.StatusCode, nil
	}

	var errResp dto.Error
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return nil, resp.StatusCode, err
	}
	return nil, resp.StatusCode, fmt.Errorf("login error: %s", errResp.Message)
}

func (s *TestSuite) registerUser(payload dto.RegisterPostRequest) (*dto.RegisterResponse, int, error) {
	client := s.server.Client()
	url := fmt.Sprintf("%s/register", s.server.URL)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	// При успешной регистрации возвращается статус 201 и RegisterResponse с полем UserID
	if resp.StatusCode == http.StatusCreated {
		var regResp dto.RegisterResponse
		if err := json.NewDecoder(resp.Body).Decode(&regResp); err != nil {
			return nil, resp.StatusCode, err
		}
		return &regResp, resp.StatusCode, nil
	}

	// При ошибке пытаемся декодировать тело как dto.Error
	var errResp dto.Error
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return nil, resp.StatusCode, err
	}
	return nil, resp.StatusCode, fmt.Errorf("registration error: %s", errResp.Message)
}

func (s *TestSuite) getToken(role string) string {
	client := s.server.Client()
	url := fmt.Sprintf("%s/dummyLogin", s.server.URL)

	payload := dto.DummyLoginPostRequest{Role: role}
	body, err := json.Marshal(payload)
	s.Require().NoError(err)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var tokenResp dto.TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	s.Require().NoError(err)
	resp.Body.Close()
	s.Require().NotEmpty(tokenResp.Token)
	return tokenResp.Token
}

func (s *TestSuite) createPvz(city, token string) (*dto.CreatePvzResponse, int, error) {
	client := s.server.Client()
	url := fmt.Sprintf("%s/pvz", s.server.URL)

	payload := dto.CreatePvzPostRequest{City: city}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		var pvzResp dto.CreatePvzResponse
		if err := json.NewDecoder(resp.Body).Decode(&pvzResp); err != nil {
			return nil, resp.StatusCode, err
		}
		return &pvzResp, resp.StatusCode, nil
	}

	var errResp dto.Error
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return nil, resp.StatusCode, err
	}
	return nil, resp.StatusCode, fmt.Errorf("pvz creation error: %s", errResp.Message)
}

func (s *TestSuite) createReception(pvzID, token string, dateTime time.Time) (*dto.CreateReceptionResponse, int, error) {
	client := s.server.Client()
	url := fmt.Sprintf("%s/receptions", s.server.URL)

	payload := dto.CreateReceptionRequest{
		PvzId:    pvzID,
		DateTime: dateTime,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		var recResp dto.CreateReceptionResponse
		if err := json.NewDecoder(resp.Body).Decode(&recResp); err != nil {
			return nil, resp.StatusCode, err
		}
		return &recResp, resp.StatusCode, nil
	}

	var errResp dto.Error
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return nil, resp.StatusCode, err
	}
	return nil, resp.StatusCode, fmt.Errorf("reception creation error: %s", errResp.Message)
}

func (s *TestSuite) addProduct(pvzID, token, prodType string) (*dto.ProductsPostResponse, int, error) {
	client := s.server.Client()
	url := fmt.Sprintf("%s/products", s.server.URL)

	payload := dto.ProductsPostRequest{
		PvzId: pvzID,
		Type:  prodType,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		var prodResp dto.ProductsPostResponse
		if err := json.NewDecoder(resp.Body).Decode(&prodResp); err != nil {
			return nil, resp.StatusCode, err
		}
		return &prodResp, resp.StatusCode, nil
	}

	var errResp dto.Error
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return nil, resp.StatusCode, err
	}
	return nil, resp.StatusCode, fmt.Errorf("add product error: %s", errResp.Message)
}

func (s *TestSuite) closeReception(pvzID, token string) (*dto.CloseReceptionResponse, int, error) {
	client := s.server.Client()
	url := fmt.Sprintf("%s/pvz/%s/close_last_reception", s.server.URL, pvzID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var closeResp dto.CloseReceptionResponse
		if err := json.NewDecoder(resp.Body).Decode(&closeResp); err != nil {
			return nil, resp.StatusCode, err
		}
		return &closeResp, resp.StatusCode, nil
	}

	var errResp dto.Error
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return nil, resp.StatusCode, err
	}
	return nil, resp.StatusCode, fmt.Errorf("close reception error: %s", errResp.Message)
}
