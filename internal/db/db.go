// Package db is the persistence layer for microapp-crm. It is the only package
// that imports bbolt; every other layer goes through the Store, receiving plain
// models, never *bolt.Tx or transaction-scoped byte slices. See
// docs/SPECIFICATIONS.md (Persistence Design) and .claude/rules/db-rules.md.
package db

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Bucket names. Defined as package-level []byte constants, never inline literals.
var (
	bucketLeads          = []byte("leads")
	bucketContacts       = []byte("contacts")
	bucketDeals          = []byte("deals")
	bucketContactByEmail = []byte("idx_contact_by_email")
	bucketDealByContact  = []byte("idx_deal_by_contact")
)

// allBuckets is the full set created at startup (see migrate).
var allBuckets = [][]byte{
	bucketLeads,
	bucketContacts,
	bucketDeals,
	bucketContactByEmail,
	bucketDealByContact,
}

// ErrNotFound is returned when a record with the requested ID does not exist.
// Match it with errors.Is, never by string.
var ErrNotFound = errors.New("not found")

// errEmptyName is the validation failure for a required name/title left blank.
var errEmptyName = errors.New("name must not be empty")

// putJSON marshals v and stores it under key in bucket b.
func putJSON(b *bolt.Bucket, key []byte, v any) error {
	encoded, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return b.Put(key, encoded)
}

// Store owns the bbolt database and exposes repository operations for all
// entities. Cross-entity use-cases (lead conversion, contact cascade-delete)
// live on the single Store so they can run inside one transaction.
type Store struct {
	db  *bolt.DB
	now func() time.Time
}

// Option configures a Store at Open time.
type Option func(*Store)

// WithClock overrides the time source (default time.Now). Intended for tests
// that need deterministic, advanceable timestamps.
func WithClock(now func() time.Time) Option {
	return func(s *Store) { s.now = now }
}

// Open opens (creating if needed) the bbolt file at path, applies the bucket
// migration, and returns a ready Store. A Timeout makes a stale lock fail fast
// instead of blocking forever — this is what enforces the single-writer /
// alternate-mode contract between the TUI and MCP surfaces.
func Open(path string, opts ...Option) (*Store, error) {
	bdb, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open bbolt at %q: %w", path, err)
	}
	s := &Store{db: bdb, now: time.Now}
	for _, o := range opts {
		o(s)
	}
	if err := s.migrate(); err != nil {
		_ = bdb.Close()
		return nil, err
	}
	return s, nil
}

// Close closes the underlying database.
func (s *Store) Close() error { return s.db.Close() }

// migrate creates every required top-level bucket idempotently in one txn.
func (s *Store) migrate() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		for _, name := range allBuckets {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return fmt.Errorf("create bucket %q: %w", name, err)
			}
		}
		return nil
	})
}

// itob encodes a surrogate ID big-endian so that byte-sorted key order matches
// numeric (creation) order.
func itob(id uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, id)
	return b
}

// btoi decodes an 8-byte big-endian ID.
func btoi(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

// normEmail normalizes an email for indexing and comparison: trimmed, lowercased.
func normEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// contactEmailIndexKey builds the composite key for idx_contact_by_email:
// normalized-email + 0x00 + big-endian contactID. The trailing ID disambiguates
// duplicate emails, and the 0x00 separator bounds the email prefix for scans.
func contactEmailIndexKey(email string, id uint64) []byte {
	e := normEmail(email)
	key := make([]byte, 0, len(e)+1+8)
	key = append(key, e...)
	key = append(key, 0x00)
	key = append(key, itob(id)...)
	return key
}

// contactEmailIndexPrefix is the scan prefix matching every contact with email.
func contactEmailIndexPrefix(email string) []byte {
	e := normEmail(email)
	prefix := make([]byte, 0, len(e)+1)
	prefix = append(prefix, e...)
	prefix = append(prefix, 0x00)
	return prefix
}
