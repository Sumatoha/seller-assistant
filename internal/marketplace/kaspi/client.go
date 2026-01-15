package kaspi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/yourusername/seller-assistant/internal/marketplace"
)

const (
	kaspiAPIBaseURL = "https://kaspi.kz/merchantcabinet/api/v1"
)

// Client implements marketplace.MarketplaceClient for Kaspi
type Client struct {
	apiKey     string
	merchantID string
	httpClient *http.Client
}

func NewClient(apiKey, merchantID string) *Client {
	return &Client{
		apiKey:     apiKey,
		merchantID: merchantID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) GetProducts() ([]marketplace.ProductData, error) {
	// Note: This is a mock implementation. Replace with actual Kaspi API endpoints
	// when the real API documentation is available
	url := fmt.Sprintf("%s/merchants/%s/products", kaspiAPIBaseURL, c.merchantID)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response struct {
		Data []struct {
			ID       string  `json:"id"`
			SKU      string  `json:"sku"`
			Name     string  `json:"name"`
			Stock    int     `json:"stock"`
			Price    float64 `json:"price"`
			Currency string  `json:"currency"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	products := make([]marketplace.ProductData, 0, len(response.Data))
	for _, p := range response.Data {
		products = append(products, marketplace.ProductData{
			ExternalID:   p.ID,
			SKU:          p.SKU,
			Name:         p.Name,
			CurrentStock: p.Stock,
			Price:        p.Price,
			Currency:     p.Currency,
		})
	}

	return products, nil
}

func (c *Client) GetProductStock(externalID string) (int, error) {
	url := fmt.Sprintf("%s/merchants/%s/products/%s/stock", kaspiAPIBaseURL, c.merchantID, externalID)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var response struct {
		Stock int `json:"stock"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Stock, nil
}

func (c *Client) GetSalesData(startDate, endDate time.Time) ([]marketplace.SalesData, error) {
	url := fmt.Sprintf("%s/merchants/%s/sales?start_date=%s&end_date=%s",
		kaspiAPIBaseURL,
		c.merchantID,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response struct {
		Data []struct {
			ProductID    string    `json:"product_id"`
			Date         time.Time `json:"date"`
			QuantitySold int       `json:"quantity_sold"`
			Revenue      float64   `json:"revenue"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	salesData := make([]marketplace.SalesData, 0, len(response.Data))
	for _, s := range response.Data {
		salesData = append(salesData, marketplace.SalesData{
			ProductExternalID: s.ProductID,
			Date:              s.Date,
			QuantitySold:      s.QuantitySold,
			Revenue:           s.Revenue,
		})
	}

	return salesData, nil
}

func (c *Client) GetReviews() ([]marketplace.ReviewData, error) {
	url := fmt.Sprintf("%s/merchants/%s/reviews", kaspiAPIBaseURL, c.merchantID)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response struct {
		Data []struct {
			ID         string    `json:"id"`
			ProductID  string    `json:"product_id"`
			AuthorName string    `json:"author_name"`
			Rating     int       `json:"rating"`
			Comment    string    `json:"comment"`
			Language   string    `json:"language"`
			CreatedAt  time.Time `json:"created_at"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	reviews := make([]marketplace.ReviewData, 0, len(response.Data))
	for _, r := range response.Data {
		reviews = append(reviews, marketplace.ReviewData{
			ExternalID: r.ID,
			ProductID:  r.ProductID,
			AuthorName: r.AuthorName,
			Rating:     r.Rating,
			Comment:    r.Comment,
			Language:   r.Language,
			CreatedAt:  r.CreatedAt,
		})
	}

	return reviews, nil
}

func (c *Client) PostReviewResponse(reviewID, response string) error {
	url := fmt.Sprintf("%s/merchants/%s/reviews/%s/response", kaspiAPIBaseURL, c.merchantID, reviewID)

	payload := map[string]string{
		"response": response,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := c.makeRequest("POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to post review response: %s", string(body))
	}

	return nil
}

func (c *Client) makeRequest(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return resp, nil
}
