package bsite

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type BSiteTemplate struct {
	ID              string
	Domain          string
	Name            string
	HostingIP       string
	HostingProvider string
	WooCommerceURL  string
	Category        string
	Products        []Product
}

type Product struct {
	ID          string
	Name        string
	Price       float64
	Category    string
	Description string
	ImageURL    string
	SKU         string
}

type SiteFactory struct {
	templatesDir string
	bSiteDomains []string
}

func NewSiteFactory(templatesDir string) *SiteFactory {
	return &SiteFactory{
		templatesDir: templatesDir,
	}
}

func (f *SiteFactory) CreateSite(ctx context.Context, template BSiteTemplate) (*BSiteTemplate, error) {
	site := &BSiteTemplate{
		ID:     uuid.New().String(),
		Domain: pickDomain(f.bSiteDomains),
		Name:   template.Name,
	}

	log.Info().
		Str("site_id", site.ID).
		Str("domain", site.Domain).
		Msg("B-site created")

	return site, nil
}

func (f *SiteFactory) SetupWooCommerce(ctx context.Context, site *BSiteTemplate, adminUser, adminPass string) error {
	log.Info().Str("domain", site.Domain).Msg("setting up WooCommerce")

	// Steps:
	// 1. Provision hosting / VPS
	// 2. Install WordPress + WooCommerce
	// 3. Install payment plugin
	// 4. Import products
	// 5. Configure SSL
	// 6. Set up social media profiles

	return nil
}

func (f *SiteFactory) InstallPaymentPlugin(ctx context.Context, site *BSiteTemplate, gateway string) error {
	log.Info().Str("gateway", gateway).Str("site", site.Domain).Msg("installing payment plugin")
	return nil
}

func (f *SiteFactory) GenerateProducts(category string, count int) []Product {
	products := make([]Product, count)
	for i := 0; i < count; i++ {
		products[i] = Product{
			ID:       uuid.New().String(),
			Name:     generateProductName(category),
			Price:    generateRealisticPrice(category),
			Category: category,
		}
	}
	return products
}

func pickDomain(domains []string) string {
	if len(domains) == 0 {
		return "example.com"
	}
	return domains[0]
}

func generateProductName(category string) string {
	names := map[string][]string{
		"electronics": {"Wireless Bluetooth Earbuds", "USB-C Hub Adapter", "Portable Phone Charger"},
		"clothing":    {"Classic Fit Cotton T-Shirt", "Slim Fit Chinos", "Lightweight Running Jacket"},
		"home":        {"Stainless Steel Water Bottle", "LED Desk Lamp", "Kitchen Utensil Set"},
	}
	if items, ok := names[category]; ok {
		return items[time.Now().Nanosecond()%len(items)]
	}
	return "Premium Quality Product"
}

func generateRealisticPrice(category string) float64 {
	prices := map[string][]float64{
		"electronics": {12.99, 19.99, 29.99, 49.99},
		"clothing":    {9.99, 15.99, 24.99, 34.99},
		"home":        {8.99, 14.99, 22.99, 39.99},
	}
	if items, ok := prices[category]; ok {
		return items[time.Now().Nanosecond()%len(items)]
	}
	return 19.99
}
