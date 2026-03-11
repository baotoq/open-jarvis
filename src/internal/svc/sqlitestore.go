package svc

import (
	"database/sql"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// SQLiteConvStore is a persistent conversation store backed by SQLite.
type SQLiteConvStore struct {
	db *sql.DB
}

// NewSQLiteConvStore creates a new SQLiteConvStore using the given *sql.DB,
// running schema migrations on startup.
func NewSQLiteConvStore(db *sql.DB) (*SQLiteConvStore, error) {
	s := &SQLiteConvStore{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

// migrate creates the schema if it does not exist.
func (s *SQLiteConvStore) migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS conversations (
    id         TEXT PRIMARY KEY,
    title      TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS messages (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id TEXT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role            TEXT NOT NULL,
    content         TEXT NOT NULL,
    position        INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_messages_conv_id ON messages(conversation_id, position);
`
	_, err := s.db.Exec(schema)
	return err
}

// Get returns all messages for the given sessionID in order, or nil if not found.
func (s *SQLiteConvStore) Get(sessionID string) []openai.ChatCompletionMessage {
	rows, err := s.db.Query(
		`SELECT role, content FROM messages WHERE conversation_id = ? ORDER BY position`,
		sessionID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var msgs []openai.ChatCompletionMessage
	for rows.Next() {
		var role, content string
		if err := rows.Scan(&role, &content); err != nil {
			return nil
		}
		msgs = append(msgs, openai.ChatCompletionMessage{Role: role, Content: content})
	}
	if err := rows.Err(); err != nil {
		return nil
	}
	return msgs
}

// Set replaces all messages for the given sessionID, upserting the conversation record.
func (s *SQLiteConvStore) Set(sessionID string, msgs []openai.ChatCompletionMessage) {
	now := time.Now().Unix()

	tx, err := s.db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Upsert conversation row
	_, err = tx.Exec(
		`INSERT INTO conversations(id, title, created_at, updated_at)
		 VALUES(?, '', ?, ?)
		 ON CONFLICT(id) DO UPDATE SET updated_at = excluded.updated_at`,
		sessionID, now, now,
	)
	if err != nil {
		return
	}

	// Delete existing messages
	_, err = tx.Exec(`DELETE FROM messages WHERE conversation_id = ?`, sessionID)
	if err != nil {
		return
	}

	// Insert all messages with position index
	for i, msg := range msgs {
		_, err = tx.Exec(
			`INSERT INTO messages(conversation_id, role, content, position) VALUES(?, ?, ?, ?)`,
			sessionID, msg.Role, msg.Content, i,
		)
		if err != nil {
			return
		}
	}

	err = tx.Commit()
}

// ListConversations returns all conversations ordered by updated_at descending.
func (s *SQLiteConvStore) ListConversations() ([]Conversation, error) {
	rows, err := s.db.Query(
		`SELECT id, title, created_at, updated_at FROM conversations ORDER BY updated_at DESC, rowid DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var convs []Conversation
	for rows.Next() {
		var c Conversation
		if err := rows.Scan(&c.ID, &c.Title, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		convs = append(convs, c)
	}
	return convs, rows.Err()
}

// GetConversation returns the conversation with the given id, or nil if not found.
func (s *SQLiteConvStore) GetConversation(id string) (*Conversation, error) {
	var c Conversation
	err := s.db.QueryRow(
		`SELECT id, title, created_at, updated_at FROM conversations WHERE id = ?`, id,
	).Scan(&c.ID, &c.Title, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// DeleteConversation deletes the conversation by id; CASCADE removes messages.
func (s *SQLiteConvStore) DeleteConversation(id string) error {
	_, err := s.db.Exec(`DELETE FROM conversations WHERE id = ?`, id)
	return err
}

// CreateConversation inserts a new conversation with the given id and title.
// Title is truncated to 50 runes.
func (s *SQLiteConvStore) CreateConversation(id, title string) error {
	runes := []rune(title)
	if len(runes) > 50 {
		runes = runes[:50]
	}
	title = string(runes)

	now := time.Now().Unix()
	_, err := s.db.Exec(
		`INSERT INTO conversations(id, title, created_at, updated_at) VALUES(?, ?, ?, ?)`,
		id, title, now, now,
	)
	return err
}
