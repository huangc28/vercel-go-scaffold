package commands

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github/huangc28/kikichoice-be/api/go/_internal/db"

	"go.uber.org/fx"
)

type CommandDAO struct {
	db db.Conn
}

type CommandDAOParams struct {
	fx.In

	DB db.Conn
}

// UserSession represents a user session from the database
type UserSession struct {
	ID          int64           `json:"id"`
	ChatID      int64           `json:"chat_id"`
	UserID      int64           `json:"user_id"`
	SessionType string          `json:"session_type"`
	State       json.RawMessage `json:"state"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	ExpiresAt   time.Time       `json:"expires_at"`
}

func NewCommandDAO(p CommandDAOParams) *CommandDAO {
	return &CommandDAO{db: p.DB}
}

// GetUserSession retrieves a user session by user_id and session_type
func (cmd *CommandDAO) GetUserSession(ctx context.Context, userID int64, sessionType string) (*UserSession, error) {
	query := `
		SELECT id, chat_id, user_id, session_type, state, created_at, updated_at, expires_at
		FROM user_sessions
		WHERE user_id = $1 AND session_type = $2 AND expires_at > NOW()
		LIMIT 1
	`

	var session UserSession
	err := cmd.db.QueryRow(ctx, query, userID, sessionType).Scan(
		&session.ID,
		&session.ChatID,
		&session.UserID,
		&session.SessionType,
		&session.State,
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.ExpiresAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No session found
		}
		return nil, err
	}

	return &session, nil
}

// CreateUserSession creates a new user session
func (cmd *CommandDAO) CreateUserSession(ctx context.Context, chatID, userID int64, sessionType string, state interface{}) error {
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO user_sessions (chat_id, user_id, session_type, state, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, session_type)
		DO UPDATE SET
			chat_id = EXCLUDED.chat_id,
			state = EXCLUDED.state,
			updated_at = NOW(),
			expires_at = EXCLUDED.expires_at
	`

	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = cmd.db.Exec(query, chatID, userID, sessionType, stateJSON, expiresAt)
	return err
}

// UpdateUserSession updates an existing user session
func (cmd *CommandDAO) UpdateUserSession(ctx context.Context, userID int64, sessionType string, state interface{}) error {
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return err
	}

	query := `
		UPDATE user_sessions
		SET state = $3, updated_at = NOW(), expires_at = $4
		WHERE user_id = $1 AND session_type = $2
	`

	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = cmd.db.Exec(query, userID, sessionType, stateJSON, expiresAt)
	return err
}

// DeleteUserSession deletes a user session
func (cmd *CommandDAO) DeleteUserSession(ctx context.Context, userID int64, sessionType string) error {
	query := `DELETE FROM user_sessions WHERE user_id = $1 AND session_type = $2`
	_, err := cmd.db.Exec(query, userID, sessionType)
	return err
}

// UpdateSession updates session state (keeping the original method signature)
func (cmd *CommandDAO) UpdateSession(ctx context.Context, userID int64, sessType string, state interface{}) error {
	return cmd.UpdateUserSession(ctx, userID, sessType, state)
}

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
