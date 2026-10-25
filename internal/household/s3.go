package household

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

var (
	ErrReadOnly           = errors.New("cached household is read-only")
	ErrStorageUnavailable = errors.New("storage unavailable")
	ErrStorageConfig      = errors.New("invalid storage configuration or household")
	ErrMultipleHouseholds = errors.New("multiple households in bucket")
	ErrUncertain          = errors.New("unable to confirm remote save")
)

type StorageConfig struct {
	Backend, Endpoint, Bucket, Region, AccessKey, SecretKey, SessionToken string
	Timeout                                                               time.Duration
}

// @impl CON-009-AC1 CON-009-AC2: static environment only, no credential/region discovery.
func StorageConfigFromEnv(get func(string) string) (StorageConfig, error) {
	c := StorageConfig{Backend: get("WISHMAS_STORAGE")}
	if c.Backend == "" {
		c.Backend = "local"
	}
	if c.Backend == "local" {
		return c, nil
	}
	if c.Backend != "s3" {
		return c, errors.New("WISHMAS_STORAGE must be local or s3")
	}
	c.Endpoint = get("WISHMAS_S3_ENDPOINT")
	c.Bucket = get("WISHMAS_S3_BUCKET")
	c.Region = get("WISHMAS_S3_REGION")
	c.AccessKey = get("WISHMAS_S3_ACCESS_KEY_ID")
	c.SecretKey = get("WISHMAS_S3_SECRET_ACCESS_KEY")
	c.SessionToken = get("WISHMAS_S3_SESSION_TOKEN")
	c.Timeout = 5 * time.Second
	if c.Region == "" {
		c.Region = "us-east-1"
	}
	if raw := get("WISHMAS_S3_TIMEOUT"); raw != "" {
		var err error
		c.Timeout, err = time.ParseDuration(raw)
		if err != nil || c.Timeout <= 0 {
			return c, errors.New("WISHMAS_S3_TIMEOUT must be a positive duration")
		}
	}
	u, err := url.Parse(c.Endpoint)
	if err != nil || u == nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return c, errors.New("WISHMAS_S3_ENDPOINT must be an HTTP(S) endpoint without credentials, query or fragment")
	}
	if strings.TrimSpace(c.Bucket) == "" || strings.ContainsAny(c.Bucket, "/\\ \t\r\n") || strings.TrimSpace(c.AccessKey) == "" || strings.TrimSpace(c.SecretKey) == "" || strings.TrimSpace(c.Region) == "" {
		return c, errors.New("S3 bucket, region and credentials are required")
	}
	return c, nil
}

type cachedHousehold struct {
	Key string `json:"key"`
	Raw []byte `json:"yaml"`
}
type S3Backend struct {
	mu            sync.Mutex
	config        StorageConfig
	dataDir, root string
	client        *s3.Client
	// Protected by mu; retained until a successful local refresh, including across redirects.
	cacheWarning bool
}

// @impl CON-009-AC2 CON-009-AC8 CON-009-AC15
func NewS3Backend(dataDir string, c StorageConfig) (*S3Backend, error) {
	// Validate explicit configurations through the same path as startup.
	env := map[string]string{"WISHMAS_STORAGE": "s3", "WISHMAS_S3_ENDPOINT": c.Endpoint, "WISHMAS_S3_BUCKET": c.Bucket, "WISHMAS_S3_REGION": c.Region, "WISHMAS_S3_ACCESS_KEY_ID": c.AccessKey, "WISHMAS_S3_SECRET_ACCESS_KEY": c.SecretKey, "WISHMAS_S3_SESSION_TOKEN": c.SessionToken}
	if c.Timeout != 0 {
		env["WISHMAS_S3_TIMEOUT"] = c.Timeout.String()
	}
	validated, err := StorageConfigFromEnv(func(k string) string { return env[k] })
	if err != nil {
		return nil, err
	}
	c = validated
	identity := sha256.Sum256([]byte(c.Endpoint + "\x00" + c.Bucket))
	root := filepath.Join(dataDir, "s3", hex.EncodeToString(identity[:]))
	cfg := aws.Config{Region: c.Region, Credentials: credentials.NewStaticCredentialsProvider(c.AccessKey, c.SecretKey, c.SessionToken), HTTPClient: &http.Client{Timeout: c.Timeout}, RetryMaxAttempts: 1, RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired, ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) { o.BaseEndpoint = aws.String(c.Endpoint); o.UsePathStyle = true })
	return &S3Backend{config: c, dataDir: dataDir, root: root, client: client}, nil
}
func (b *S3Backend) operation() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), b.config.Timeout)
}
func classifyS3(err error) error {
	var response *smithyhttp.ResponseError
	if errors.As(err, &response) {
		code := response.HTTPStatusCode()
		if code >= 500 || code == 408 || code == 429 {
			return ErrStorageUnavailable
		}
		if code == 404 {
			return os.ErrNotExist
		}
		return ErrStorageConfig
	}
	// Transport/deadline failures are eligible for fallback. Never expose SDK errors/credentials to the UI.
	return ErrStorageUnavailable
}
func (b *S3Backend) keys() ([]string, error) {
	ctx, cancel := b.operation()
	defer cancel()
	keys := []string{}
	pager := s3.NewListObjectsV2Paginator(b.client, &s3.ListObjectsV2Input{Bucket: aws.String(b.config.Bucket)})
	for pager.HasMorePages() {
		out, err := pager.NextPage(ctx)
		if err != nil {
			e := classifyS3(err)
			if errors.Is(e, os.ErrNotExist) {
				e = ErrStorageConfig
			}
			return nil, e
		}
		for _, o := range out.Contents {
			k := aws.ToString(o.Key)
			parts := strings.Split(k, "/")
			if len(parts) == 2 && parts[0] != "" && parts[1] == "household.yaml" {
				keys = append(keys, k)
			}
		}
	}
	return keys, nil
}
func (b *S3Backend) get(key string) ([]byte, error) {
	ctx, cancel := b.operation()
	defer cancel()
	out, err := b.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(b.config.Bucket), Key: aws.String(key)})
	if err != nil {
		return nil, classifyS3(err)
	}
	defer out.Body.Close()
	raw, err := io.ReadAll(out.Body)
	if err != nil {
		return nil, ErrStorageUnavailable
	}
	return raw, nil
}
func (b *S3Backend) put(key string, raw []byte) error {
	ctx, cancel := b.operation()
	defer cancel()
	_, err := b.client.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(b.config.Bucket), Key: aws.String(key), Body: bytes.NewReader(raw), ContentType: aws.String("application/yaml")})
	if err != nil {
		if errors.Is(classifyS3(err), ErrStorageUnavailable) {
			return ErrUncertain
		}
		return ErrStorageConfig
	}
	return nil
}
func revisionOf(raw []byte) string { sum := sha256.Sum256(raw); return hex.EncodeToString(sum[:]) }
func validRemote(key string, raw []byte) (Household, error) {
	h, err := Decode(raw)
	if err != nil || key != h.Name+"/household.yaml" {
		return Household{}, ErrStorageConfig
	}
	return h, nil
}
func (b *S3Backend) store(key string) *Store {
	identity := strings.TrimSuffix(key, "/household.yaml")
	return &Store{path: filepath.Join(b.root, "households", identity, "household.yaml"), mu: &b.mu, remote: b, key: key}
}
func (b *S3Backend) cachePath() string { return filepath.Join(b.root, "fallback.json") }
func (b *S3Backend) saveCache(key string, raw []byte) {
	encoded, _ := json.Marshal(cachedHousehold{Key: key, Raw: raw})
	b.cacheWarning = atomicPrivate(b.cachePath(), encoded) != nil
}
func (b *S3Backend) fallback() (*Household, *Store, string, error) {
	raw, err := readRegular(b.cachePath())
	if err != nil {
		return nil, nil, "", ErrStorageUnavailable
	}
	var c cachedHousehold
	if json.Unmarshal(raw, &c) != nil {
		return nil, nil, "", ErrStorageUnavailable
	}
	h, err := validRemote(c.Key, c.Raw)
	if err != nil {
		return nil, nil, "", ErrStorageUnavailable
	}
	s := b.store(c.Key)
	s.readOnly = true
	return &h, s, revisionOf(c.Raw), nil
}

// @impl CON-009-AC3 CON-009-AC4 CON-009-AC5 CON-009-AC6 CON-009-AC8 CON-009-AC9 CON-009-AC10 CON-009-AC11 CON-009-AC12
func (b *S3Backend) Load() (*Household, *Store, string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.load()
}
func (b *S3Backend) load() (*Household, *Store, string, error) {
	h, s, rev, err := b.loadRaw()
	if err == nil && h != nil && !s.readOnly {
		h, rev = b.reconcile(h, s, rev, time.Now(), false)
	}
	return h, s, rev, err
}
func (b *S3Backend) loadRaw() (*Household, *Store, string, error) {
	keys, err := b.keys()
	if errors.Is(err, ErrStorageUnavailable) {
		return b.fallback()
	}
	if err != nil {
		return nil, nil, "", err
	}
	if len(keys) > 1 {
		return nil, nil, "", ErrMultipleHouseholds
	}
	if len(keys) == 0 {
		return b.missing()
	}
	raw, err := b.get(keys[0])
	if errors.Is(err, os.ErrNotExist) {
		return b.missing()
	}
	if errors.Is(err, ErrStorageUnavailable) {
		h, s, rev, e := b.fallback()
		if e == nil && s.key != keys[0] {
			return nil, nil, "", ErrStorageUnavailable
		}
		return h, s, rev, e
	}
	if err != nil {
		return nil, nil, "", err
	}
	h, err := validRemote(keys[0], raw)
	if err != nil {
		return nil, nil, "", err
	}
	b.saveCache(keys[0], raw)
	return &h, b.store(keys[0]), revisionOf(raw), nil
}
func (b *S3Backend) missing() (*Household, *Store, string, error) {
	// A durable tombstone prevents resurrecting a deleted household after restart.
	if err := atomicPrivate(b.cachePath(), []byte(`{}`)); err != nil {
		return nil, nil, "", ErrStorageConfig
	}
	b.cacheWarning = false
	return nil, nil, "", nil
}

// @impl CON-009-AC3 CON-009-AC6 CON-009-AC7 CON-009-AC14 CON-009-AC16
func (b *S3Backend) Create(h Household) (*Store, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	keys, err := b.keys()
	if err != nil {
		return nil, err
	}
	if len(keys) != 0 {
		return nil, ErrStale
	}
	raw, err := Encode(h)
	if err != nil {
		return nil, err
	}
	key := h.Name + "/household.yaml"
	if err = b.put(key, raw); err != nil {
		return nil, err
	}
	b.saveCache(key, raw)
	return b.store(key), nil
}
func (s *Store) ReadOnly() bool { s.mu.Lock(); defer s.mu.Unlock(); return s.readOnly }
func (s *Store) CacheWarning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.remote != nil && s.remote.cacheWarning
}
func (s *Store) IsRemote() bool { return s.remote != nil }
func (s *Store) remoteLoad() (Household, string, error) {
	h, st, rev, err := s.remote.loadRaw()
	if err != nil {
		return Household{}, "", err
	}
	if h == nil || st.key != s.key {
		return Household{}, "", ErrStale
	}
	s.readOnly = st.readOnly
	return *h, rev, nil
}

// @impl CON-009-AC7 CON-009-AC13 CON-009-AC14
func (s *Store) remoteWrite(raw []byte, create bool, revision string) error {
	h, st, rev, err := s.remote.loadRaw()
	if err != nil {
		return err
	}
	if h == nil || st.key != s.key {
		return ErrStale
	}
	if st.readOnly {
		return ErrReadOnly
	}
	if create || rev != revision {
		return ErrStale
	}
	if err = s.remote.put(s.key, raw); err != nil {
		return err
	}
	s.remote.saveCache(s.key, raw)
	return nil
}
func readRegular(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("not a regular file")
	}
	return os.ReadFile(path)
}
func atomicPrivate(path string, raw []byte) error {
	parent := filepath.Dir(path)
	if err := os.MkdirAll(parent, 0700); err != nil {
		return err
	}
	if info, err := os.Lstat(path); err == nil && !info.Mode().IsRegular() {
		return errors.New("invalid state file")
	}
	file, err := os.CreateTemp(parent, ".state-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if err = file.Chmod(0600); err == nil {
		_, err = file.Write(raw)
	}
	if err == nil {
		err = file.Sync()
	}
	if e := file.Close(); err == nil {
		err = e
	}
	if err != nil {
		return err
	}
	if err = os.Rename(file.Name(), path); err != nil {
		return err
	}
	dir, err := os.Open(parent)
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

// LoadForMaintenance refreshes authority without bypassing retry deadlines.
func (b *S3Backend) LoadForMaintenance() (*Household, *Store, string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.loadRaw()
}
