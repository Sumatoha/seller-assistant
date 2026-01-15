package marketplace

import "time"

// MarketplaceClient defines the interface for marketplace API integrations
type MarketplaceClient interface {
	// GetProducts fetches all products from the marketplace
	GetProducts() ([]ProductData, error)

	// GetProductStock fetches current stock for a specific product
	GetProductStock(externalID string) (int, error)

	// GetSalesData fetches sales data for a date range
	GetSalesData(startDate, endDate time.Time) ([]SalesData, error)

	// GetReviews fetches new reviews
	GetReviews() ([]ReviewData, error)

	// PostReviewResponse posts a response to a review
	PostReviewResponse(reviewID, response string) error
}

// ProductData represents product information from marketplace
type ProductData struct {
	ExternalID   string
	SKU          string
	Name         string
	CurrentStock int
	Price        float64
	Currency     string
}

// SalesData represents sales information from marketplace
type SalesData struct {
	ProductExternalID string
	Date              time.Time
	QuantitySold      int
	Revenue           float64
}

// ReviewData represents review information from marketplace
type ReviewData struct {
	ExternalID string
	ProductID  string
	AuthorName string
	Rating     int
	Comment    string
	Language   string
	CreatedAt  time.Time
}
