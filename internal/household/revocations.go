package household

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"time"
)

type retryRecord struct {
	Attempts int       `json:"attempts"`
	Next     time.Time `json:"next"`
}

func RetryDelay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	if attempt > 5 {
		attempt = 5
	}
	return []time.Duration{time.Minute, 2 * time.Minute, 4 * time.Minute, 8 * time.Minute, 16 * time.Minute, 30 * time.Minute}[attempt]
}
func (s *Store) revocationPath() string {
	return filepath.Join(filepath.Dir(s.path), "revocations.json")
}
func (s *Store) revocations() (map[string]retryRecord, error) {
	raw, err := readRegular(s.revocationPath())
	if os.IsNotExist(err) {
		return map[string]retryRecord{}, nil
	}
	if err != nil {
		return nil, err
	}
	var records map[string]retryRecord
	if err = json.Unmarshal(raw, &records); err != nil || records == nil {
		return nil, errors.New("invalid revocation state")
	}
	for hash := range records {
		if !validHash(hash) {
			return nil, errors.New("invalid revocation hash")
		}
	}
	return records, nil
}
func validHash(s string) bool { b, err := hex.DecodeString(s); return err == nil && len(b) == 32 }
func (s *Store) saveRevocations(records map[string]retryRecord) error {
	raw, err := json.Marshal(records)
	if err != nil {
		return err
	}
	return atomicPrivate(s.revocationPath(), raw)
}

// @impl ACC-008-AC4 ACC-008-AC5 ACC-008-AC6 ACC-008-AC10
func (s *Store) RevokeSession(hash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.remote == nil || !validHash(hash) {
		return errors.New("invalid pending revocation")
	}
	records, err := s.revocations()
	if err != nil {
		return err
	}
	records[hash] = retryRecord{}
	return s.saveRevocations(records)
}

// Corrupt/unreadable deny state fails closed; never resurrect a revoked session.
func (s *Store) SessionRevoked(hash string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.remote == nil {
		return false
	}
	records, err := s.revocations()
	if err != nil {
		return true
	}
	_, found := records[hash]
	return found
}

// @impl ACC-008-AC7 ACC-008-AC8 ACC-008-AC9
// Caller holds backend lock. Return the confirmed snapshot, whether or not a
// session-removal write succeeds. Durable denials remain authoritative locally.
func (b *S3Backend) reconcile(h *Household, s *Store, rev string, now time.Time, dueOnly bool) (*Household, string) {
	records, err := s.revocations()
	if err != nil || len(records) == 0 {
		return h, rev
	}
	due := map[string]bool{}
	for hash, record := range records {
		if !dueOnly || !now.Before(record.Next) {
			due[hash] = true
		}
	}
	if len(due) == 0 {
		return h, rev
	}
	current := *h
	current.Sessions = slices.Clone(h.Sessions)
	current.Sessions = slices.DeleteFunc(current.Sessions, func(session Session) bool { return due[session.TokenHash] })
	confirmed := len(current.Sessions) == len(h.Sessions)
	if !confirmed {
		raw, e := Encode(current)
		if e == nil {
			e = b.put(s.key, raw)
		}
		if e == nil {
			b.saveCache(s.key, raw)
			h = &current
			rev = revisionOf(raw)
			confirmed = true
		}
	}
	// A successful remote removal alone is insufficient: an older local fallback
	// may still contain the session. Retain its denial until fallback is refreshed.
	for hash := range due {
		record := records[hash]
		if confirmed && !b.cacheWarning {
			delete(records, hash)
		} else {
			record.Next = now.Add(RetryDelay(record.Attempts))
			if record.Attempts < 5 {
				record.Attempts++
			}
			records[hash] = record
		}
	}
	_ = s.saveRevocations(records) // Failure retains the previous durable deny record.
	return h, rev
}
func deferRetry(record retryRecord, now time.Time) retryRecord {
	record.Next = now.Add(RetryDelay(record.Attempts))
	if record.Attempts < 5 {
		record.Attempts++
	}
	return record
}
func hasDue(records map[string]retryRecord, now time.Time) bool {
	for _, record := range records {
		if !now.Before(record.Next) {
			return true
		}
	}
	return false
}

// @impl ACC-008-AC8 WSH-008-AC10: idle ticks inspect local deadlines only.
func (b *S3Backend) Maintenance(now time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	entries, err := os.ReadDir(filepath.Join(b.root, "households"))
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() || !validName(entry.Name(), 3, 50, false) {
			continue
		}
		s := b.store(entry.Name() + "/household.yaml")
		records, e := s.revocations()
		if e == nil && hasDue(records, now) {
			h, st, rev, e := b.loadRaw()
			if e == nil && h != nil && st.key == s.key && !st.readOnly {
				b.reconcile(h, s, rev, now, true)
			} else {
				for hash, record := range records {
					if !now.Before(record.Next) {
						records[hash] = deferRetry(record, now)
					}
				}
				_ = s.saveRevocations(records)
			}
		}
		cleanup, e := s.cleanupRecords()
		if e == nil && hasDue(cleanup, now) {
			_ = s.cleanupPictures(now, os.Remove)
		}
	}
}
