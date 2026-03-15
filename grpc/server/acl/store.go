package acl

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"path"
	"time"

	_ "modernc.org/sqlite"
)

// Store provides SQLite-backed storage for per-method ACL grants,
// client registrations, and pending access requests.
type Store struct {
	db *sql.DB
}

// ClientInfo describes a registered client.
type ClientInfo struct {
	ClientID     string
	CertPEM      string
	Fingerprint  string
	RegisteredAt time.Time
}

// GrantInfo describes a method-access grant for a client.
type GrantInfo struct {
	ClientID      string
	MethodPattern string
	GrantedAt     time.Time
	GrantedBy     string
}

// RequestInfo describes a pending (or resolved) access request.
type RequestInfo struct {
	ID          int64
	ClientID    string
	Methods     []string
	RequestedAt time.Time
	Status      string
}

// OpenStore opens (or creates) a SQLite database at dbPath and
// ensures the required tables exist.
func OpenStore(dbPath string) (_ *Store, _err error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite db: %w", err)
	}
	defer func() {
		if _err != nil {
			_ = db.Close()
		}
	}()

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("creating tables: %w", err)
	}

	return &Store{db: db}, nil
}

func createTables(db *sql.DB) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS clients (
	id            INTEGER PRIMARY KEY,
	client_id     TEXT UNIQUE NOT NULL,
	cert_pem      TEXT NOT NULL,
	fingerprint   TEXT NOT NULL,
	registered_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS grants (
	id             INTEGER PRIMARY KEY,
	client_id      TEXT NOT NULL,
	method_pattern TEXT NOT NULL,
	granted_at     TEXT NOT NULL,
	granted_by     TEXT NOT NULL,
	UNIQUE(client_id, method_pattern)
);

CREATE TABLE IF NOT EXISTS pending_requests (
	id           INTEGER PRIMARY KEY,
	client_id    TEXT NOT NULL,
	methods      TEXT NOT NULL,
	requested_at TEXT NOT NULL,
	status       TEXT NOT NULL DEFAULT 'pending'
);`

	_, err := db.Exec(ddl)
	return err
}

// Close closes the underlying database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// RegisterClient inserts a new client record.
func (s *Store) RegisterClient(
	clientID string,
	certPEM string,
	fingerprint string,
) error {
	_, err := s.db.Exec(
		`INSERT INTO clients (client_id, cert_pem, fingerprint, registered_at) VALUES (?, ?, ?, ?)`,
		clientID, certPEM, fingerprint, time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("inserting client %q: %w", clientID, err)
	}
	return nil
}

// ListClients returns all registered clients.
func (s *Store) ListClients() ([]ClientInfo, error) {
	rows, err := s.db.Query(`SELECT client_id, cert_pem, fingerprint, registered_at FROM clients`)
	if err != nil {
		return nil, fmt.Errorf("querying clients: %w", err)
	}
	defer rows.Close()

	var clients []ClientInfo
	for rows.Next() {
		var c ClientInfo
		var ts string
		if err := rows.Scan(&c.ClientID, &c.CertPEM, &c.Fingerprint, &ts); err != nil {
			return nil, fmt.Errorf("scanning client row: %w", err)
		}
		c.RegisteredAt, err = time.Parse(time.RFC3339, ts)
		if err != nil {
			return nil, fmt.Errorf("parsing registered_at %q: %w", ts, err)
		}
		clients = append(clients, c)
	}
	return clients, rows.Err()
}

// Grant inserts an access grant for the given client and method pattern.
// If the exact (clientID, methodPattern) pair already exists, the call
// is silently ignored (INSERT OR IGNORE).
func (s *Store) Grant(
	clientID string,
	methodPattern string,
	grantedBy string,
) error {
	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO grants (client_id, method_pattern, granted_at, granted_by) VALUES (?, ?, ?, ?)`,
		clientID, methodPattern, time.Now().UTC().Format(time.RFC3339), grantedBy,
	)
	if err != nil {
		return fmt.Errorf("granting %q to %q: %w", methodPattern, clientID, err)
	}
	return nil
}

// RevokeClient deletes the client record and all associated grants.
// Both operations run in a single transaction.
func (s *Store) RevokeClient(clientID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.Exec(`DELETE FROM grants WHERE client_id = ?`, clientID); err != nil {
		return fmt.Errorf("deleting grants for client %q: %w", clientID, err)
	}
	if _, err := tx.Exec(`DELETE FROM clients WHERE client_id = ?`, clientID); err != nil {
		return fmt.Errorf("deleting client %q: %w", clientID, err)
	}

	return tx.Commit()
}

// Revoke removes the grant for the given client and method pattern.
func (s *Store) Revoke(
	clientID string,
	methodPattern string,
) error {
	_, err := s.db.Exec(
		`DELETE FROM grants WHERE client_id = ? AND method_pattern = ?`,
		clientID, methodPattern,
	)
	if err != nil {
		return fmt.Errorf("revoking %q from %q: %w", methodPattern, clientID, err)
	}
	return nil
}

// ListGrants returns grants for the given client. If clientID is empty,
// all grants are returned.
func (s *Store) ListGrants(clientID string) ([]GrantInfo, error) {
	var (
		rows *sql.Rows
		err  error
	)
	switch {
	case clientID == "":
		rows, err = s.db.Query(`SELECT client_id, method_pattern, granted_at, granted_by FROM grants`)
	default:
		rows, err = s.db.Query(
			`SELECT client_id, method_pattern, granted_at, granted_by FROM grants WHERE client_id = ?`,
			clientID,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("querying grants: %w", err)
	}
	defer rows.Close()

	var grants []GrantInfo
	for rows.Next() {
		var g GrantInfo
		var ts string
		if err := rows.Scan(&g.ClientID, &g.MethodPattern, &ts, &g.GrantedBy); err != nil {
			return nil, fmt.Errorf("scanning grant row: %w", err)
		}
		g.GrantedAt, err = time.Parse(time.RFC3339, ts)
		if err != nil {
			return nil, fmt.Errorf("parsing granted_at %q: %w", ts, err)
		}
		grants = append(grants, g)
	}
	return grants, rows.Err()
}

// IsMethodAllowed checks whether any grant for the given client matches
// the full gRPC method name. Matching rules:
//   - Exact match: pattern equals fullMethod
//   - Service wildcard: "/service.Name/*" matches any method in that service
//   - Global wildcard: "/*" matches everything
//
// Uses path.Match for glob matching.
func (s *Store) IsMethodAllowed(
	clientID string,
	fullMethod string,
) (bool, error) {
	rows, err := s.db.Query(
		`SELECT method_pattern FROM grants WHERE client_id = ?`,
		clientID,
	)
	if err != nil {
		return false, fmt.Errorf("querying grants for %q: %w", clientID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var pattern string
		if err := rows.Scan(&pattern); err != nil {
			return false, fmt.Errorf("scanning grant pattern: %w", err)
		}
		if matchesMethod(pattern, fullMethod) {
			return true, nil
		}
	}
	return false, rows.Err()
}

// matchesMethod checks whether pattern matches fullMethod.
//
// Because path.Match treats "/" as a separator and "*" does not cross
// segment boundaries, the global wildcard "/*" is handled explicitly:
// it matches any method that starts with "/". Service wildcards like
// "/service.Name/*" work naturally with path.Match because the method
// part is a single segment.
func matchesMethod(pattern, fullMethod string) bool {
	if pattern == fullMethod {
		return true
	}

	// Global wildcard: "/*" matches any gRPC method (they all start with "/").
	if pattern == "/*" {
		return len(fullMethod) > 0 && fullMethod[0] == '/'
	}

	matched, err := path.Match(pattern, fullMethod)
	if err != nil {
		return false
	}
	return matched
}

// CreateRequest stores a new pending access request. The methods slice
// is serialized as a JSON array. Returns the request row ID.
func (s *Store) CreateRequest(
	clientID string,
	methods []string,
) (int64, error) {
	methodsJSON, err := json.Marshal(methods)
	if err != nil {
		return 0, fmt.Errorf("marshaling methods: %w", err)
	}

	result, err := s.db.Exec(
		`INSERT INTO pending_requests (client_id, methods, requested_at, status) VALUES (?, ?, ?, 'pending')`,
		clientID, string(methodsJSON), time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return 0, fmt.Errorf("inserting request for %q: %w", clientID, err)
	}
	return result.LastInsertId()
}

// ListPendingRequests returns all requests with status='pending'.
func (s *Store) ListPendingRequests() ([]RequestInfo, error) {
	rows, err := s.db.Query(
		`SELECT id, client_id, methods, requested_at, status FROM pending_requests WHERE status = 'pending'`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying pending requests: %w", err)
	}
	defer rows.Close()

	var requests []RequestInfo
	for rows.Next() {
		r, err := scanRequestRow(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, r)
	}
	return requests, rows.Err()
}

func scanRequestRow(rows *sql.Rows) (RequestInfo, error) {
	var r RequestInfo
	var methodsJSON, ts string
	if err := rows.Scan(&r.ID, &r.ClientID, &methodsJSON, &ts, &r.Status); err != nil {
		return RequestInfo{}, fmt.Errorf("scanning request row: %w", err)
	}

	var err error
	r.RequestedAt, err = time.Parse(time.RFC3339, ts)
	if err != nil {
		return RequestInfo{}, fmt.Errorf("parsing requested_at %q: %w", ts, err)
	}

	if err := json.Unmarshal([]byte(methodsJSON), &r.Methods); err != nil {
		return RequestInfo{}, fmt.Errorf("unmarshaling methods %q: %w", methodsJSON, err)
	}
	return r, nil
}

// ApproveRequest marks the request as 'approved' and creates grants
// for each method in the request. Both operations run in a single
// transaction.
func (s *Store) ApproveRequest(
	requestID int64,
	approvedBy string,
) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var clientID, methodsJSON string
	err = tx.QueryRow(
		`SELECT client_id, methods FROM pending_requests WHERE id = ?`,
		requestID,
	).Scan(&clientID, &methodsJSON)
	if err != nil {
		return fmt.Errorf("looking up request %d: %w", requestID, err)
	}

	var methods []string
	if err := json.Unmarshal([]byte(methodsJSON), &methods); err != nil {
		return fmt.Errorf("unmarshaling methods for request %d: %w", requestID, err)
	}

	_, err = tx.Exec(
		`UPDATE pending_requests SET status = 'approved' WHERE id = ?`,
		requestID,
	)
	if err != nil {
		return fmt.Errorf("updating request %d status: %w", requestID, err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	for _, method := range methods {
		_, err = tx.Exec(
			`INSERT OR IGNORE INTO grants (client_id, method_pattern, granted_at, granted_by) VALUES (?, ?, ?, ?)`,
			clientID, method, now, approvedBy,
		)
		if err != nil {
			return fmt.Errorf("granting %q to %q: %w", method, clientID, err)
		}
	}

	return tx.Commit()
}

// DenyRequest marks the request as 'denied' without creating any grants.
func (s *Store) DenyRequest(requestID int64) error {
	_, err := s.db.Exec(
		`UPDATE pending_requests SET status = 'denied' WHERE id = ?`,
		requestID,
	)
	if err != nil {
		return fmt.Errorf("denying request %d: %w", requestID, err)
	}
	return nil
}
