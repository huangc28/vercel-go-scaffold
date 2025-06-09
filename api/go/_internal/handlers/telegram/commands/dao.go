package commands

import (
	"context"
	"encoding/json"
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
func (cmd *CommandDAO) GetUserSession(ctx context.Context, userID, chatID int64, sessionType string) (*db.UserSession, error) {
	query := `
		SELECT
			id,
			chat_id,
			user_id,
			session_type,
			state,
			created_at,
			updated_at,
			expires_at
		FROM user_sessions
		WHERE
			user_id = $1 AND
			chat_id = $2 AND
			session_type = $3 AND
			expires_at > NOW()
		LIMIT 1
	`

	var session db.UserSession
	if err := cmd.db.QueryRowx(query, userID, chatID, sessionType).StructScan(&session); err != nil {
		return nil, err
	}

	return &session, nil
}

// CreateUserSession creates a new user session
func (cmd *CommandDAO) UpsertUserSession(ctx context.Context, chatID, userID int64, sessionType string, state interface{}) error {
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
	_, err = cmd.db.Exec(query, chatID, userID, sessionType, string(stateJSON), expiresAt)
	return err
}

// DeleteUserSession deletes a user session
func (cmd *CommandDAO) DeleteUserSession(ctx context.Context, userID int64, sessionType string) error {
	query := `DELETE FROM user_sessions WHERE user_id = $1 AND session_type = $2`
	_, err := cmd.db.Exec(query, userID, sessionType)
	return err
}
