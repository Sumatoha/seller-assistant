package kaspi

import (
	"math/rand"
	"time"
)

// CompetitorPrice представляет цену конкурента
type CompetitorPrice struct {
	SellerName string  `json:"seller_name"`
	Price      float64 `json:"price"`
}

// GetCompetitorPrices получает цены конкурентов для товара (MOCK)
// В реальной реализации здесь будет запрос к API Kaspi
func (c *Client) GetCompetitorPrices(productExternalID string) ([]CompetitorPrice, error) {
	// MOCK: Генерируем случайные цены конкурентов
	// В реальности здесь будет HTTP запрос к Kaspi API

	time.Sleep(100 * time.Millisecond) // Имитация задержки сети

	numCompetitors := rand.Intn(5) + 2 // От 2 до 6 конкурентов
	prices := make([]CompetitorPrice, numCompetitors)

	basePrice := 10000.0 + rand.Float64()*50000.0 // Базовая цена от 10к до 60к

	for i := 0; i < numCompetitors; i++ {
		variation := (rand.Float64() - 0.5) * 0.2 // Вариация ±10%
		price := basePrice * (1 + variation)

		prices[i] = CompetitorPrice{
			SellerName: generateSellerName(i),
			Price:      roundToTenge(price),
		}
	}

	return prices, nil
}

// UpdateProductPrice обновляет цену товара на Kaspi (MOCK)
func (c *Client) UpdateProductPrice(productExternalID string, newPrice float64) error {
	// MOCK: В реальности здесь будет HTTP PUT/PATCH запрос к Kaspi API
	// для обновления цены товара

	time.Sleep(100 * time.Millisecond) // Имитация задержки сети

	// Логируем (в реальности здесь будет реальный запрос)
	// log.Printf("Updating price for product %s to %.2f", productExternalID, newPrice)

	return nil
}

// GetMinCompetitorPrice возвращает минимальную цену среди конкурентов
func GetMinCompetitorPrice(prices []CompetitorPrice) float64 {
	if len(prices) == 0 {
		return 0
	}

	minPrice := prices[0].Price
	for _, p := range prices {
		if p.Price < minPrice {
			minPrice = p.Price
		}
	}

	return minPrice
}

// Helper functions

func generateSellerName(index int) string {
	names := []string{
		"TechnoShop KZ",
		"Mega Store",
		"Digital World",
		"Smart Electronics",
		"Best Price KZ",
		"Tech Master",
		"Gadget Paradise",
		"Kazakhstan Electronics",
	}

	if index < len(names) {
		return names[index]
	}

	return "Seller " + string(rune('A'+index))
}

func roundToTenge(price float64) float64 {
	return float64(int(price))
}
