package svc

import (
	"database/sql"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// SearchResult holds a conversation match returned by SearchConversations.
type SearchResult struct {
	ConversationID string
	Title          string
	UpdatedAt      int64
	Snippet        string
}

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
	if _, err := s.db.Exec(schema); err != nil {
		return err
	}

	// FTS5 virtual table for full-text search over message content.
	// Split into separate Exec calls because the trigger BEGIN...END syntax
	// does not work reliably in a single multi-statement Exec.
	ftsStatements := []string{
		`CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
			content,
			content='messages',
			content_rowid='id'
		)`,
		`CREATE TRIGGER IF NOT EXISTS messages_ai AFTER INSERT ON messages BEGIN
			INSERT INTO messages_fts(rowid, content) VALUES (new.id, new.content);
		END`,
		`CREATE TRIGGER IF NOT EXISTS messages_ad AFTER DELETE ON messages BEGIN
			INSERT INTO messages_fts(messages_fts, rowid, content)
				VALUES ('delete', old.id, old.content);
		END`,
		`CREATE TRIGGER IF NOT EXISTS messages_au AFTER UPDATE ON messages BEGIN
			INSERT INTO messages_fts(messages_fts, rowid, content)
				VALUES ('delete', old.id, old.content);
			INSERT INTO messages_fts(rowid, content) VALUES (new.id, new.content);
		END`,
		// Initial populate: rebuild the FTS index from the content table.
		// For FTS5 content tables, SELECT rowid FROM messages_fts reflects the
		// underlying table even before explicit indexing, so the NOT IN guard
		// does not work. Using 'rebuild' forces a full re-index idempotently.
		`INSERT INTO messages_fts(messages_fts) VALUES('rebuild')`,
	}
	for _, stmt := range ftsStatements {
		if _, err := s.db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
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

// SanitizeFTSQuery wraps user input in double-quotes and escapes internal
// double-quotes for safe use in an FTS5 MATCH expression.
// Returns "" for blank input, which callers should treat as "no search".
func SanitizeFTSQuery(q string) string {
	q = strings.TrimSpace(q)
	if q == "" {
		return ""
	}
	// Escape internal double-quotes by doubling them, then wrap in quotes.
	escaped := strings.ReplaceAll(q, `"`, `""`)
	return `"` + escaped + `"`
}

const searchSQL = `
SELECT DISTINCT m.conversation_id, c.title, c.updated_at,
       snippet(messages_fts, 0, '<b>', '</b>', '...', 20) AS snippet
FROM messages_fts
JOIN messages m ON messages_fts.rowid = m.id
JOIN conversations c ON m.conversation_id = c.id
WHERE messages_fts MATCH ?
ORDER BY rank
LIMIT 20`

// SearchConversations returns up to 20 conversations whose messages match
// the query string, ordered by FTS5 relevance rank.
func (s *SQLiteConvStore) SearchConversations(query string) ([]SearchResult, error) {
	sanitized := SanitizeFTSQuery(query)
	if sanitized == "" {
		return nil, nil
	}

	rows, err := s.db.Query(searchSQL, sanitized)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.ConversationID, &r.Title, &r.UpdatedAt, &r.Snippet); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}
