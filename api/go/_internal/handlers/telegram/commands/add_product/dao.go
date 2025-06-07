package add_product

import (
	"context"
	"fmt"

	"github/huangc28/kikichoice-be/api/go/_internal/db"

	"go.uber.org/fx"
)

// ProductDAO handles product-related database operations
type ProductDAO struct {
	db db.Conn
}

type ProductDAOParams struct {
	fx.In

	DB db.Conn
}

func NewProductDAO(p ProductDAOParams) *ProductDAO {
	return &ProductDAO{db: p.DB}
}

// SaveProduct saves the product to the database using raw SQL
func (p *ProductDAO) SaveProduct(ctx context.Context, state *UserState) error {
	// Create product using raw SQL
	query := `
		INSERT INTO products (sku, name, price, category, stock_count, description)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var productID int64
	err := p.db.QueryRow(query,
		state.Product.SKU,
		state.Product.Name,
		state.Product.Price,
		state.Product.Category,
		state.Product.Stock,
		state.Product.Description,
	).Scan(&productID)

	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	// Create product specs
	for i, spec := range state.Specs {
		// Assume spec format is "name:value"
		specQuery := `
			INSERT INTO product_specs (product_id, spec_name, spec_value, sort_order)
			VALUES ($1, $2, $3, $4)
		`
		// Simple parsing - you might want to improve this
		specName := spec
		specValue := ""
		if len(spec) > 0 {
			specName = spec
			specValue = spec // For now, store the whole string as both name and value
		}

		_, err := p.db.Exec(specQuery, productID, specName, specValue, i)
		if err != nil {
			return fmt.Errorf("failed to create product spec: %w", err)
		}
	}

	// Create product images
	for i, fileID := range state.ImageFileIDs {
		imageQuery := `
			INSERT INTO product_images (product_id, url, alt_text, is_primary, sort_order)
			VALUES ($1, $2, $3, $4, $5)
		`

		isPrimary := i == 0 // First image is primary
		altText := fmt.Sprintf("%s image %d", state.Product.Name, i+1)
		url := fmt.Sprintf("telegram_file://%s", fileID)

		_, err := p.db.Exec(imageQuery, productID, url, altText, isPrimary, i)
		if err != nil {
			return fmt.Errorf("failed to create product image: %w", err)
		}
	}

	return nil
}
