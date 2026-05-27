package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/AnimusHQ/news/internal/artifacts"
)

var (
	ErrImmutableConflict = errors.New("artifact ref is immutable")
	ErrNotFound          = errors.New("storage record not found")
	ErrValidationFailed  = errors.New("artifact validation failed")
)

// ArtifactStore stores immutable, content-addressed episode artifacts.
type ArtifactStore interface {
	PutArtifact(ctx context.Context, input PutArtifactInput) (ArtifactRef, error)
	GetArtifact(ctx context.Context, ref ArtifactRef) ([]byte, error)
	ListArtifacts(ctx context.Context, episodeID string) ([]ArtifactRef, error)
}

// EpisodeRepository stores durable episode state and links to artifact refs.
type EpisodeRepository interface {
	SaveEpisode(ctx context.Context, record EpisodeRecord) error
	GetEpisode(ctx context.Context, episodeID string) (EpisodeRecord, error)
	AddArtifactRef(ctx context.Context, episodeID string, ref ArtifactRef) error
}

// PutArtifactInput describes one artifact write request.
type PutArtifactInput struct {
	EpisodeID         string            `json:"episode_id"`
	ArtifactName      string            `json:"artifact_name"`
	Content           []byte            `json:"-"`
	ValidateCanonical bool              `json:"validate_canonical"`
	CreatedAt         time.Time         `json:"created_at,omitempty"`
	Metadata          map[string]string `json:"metadata,omitempty"`
}

// ArtifactRef is the durable reference returned by an ArtifactStore.
type ArtifactRef struct {
	EpisodeID    string            `json:"episode_id"`
	ArtifactName string            `json:"artifact_name"`
	ContentHash  string            `json:"content_hash"`
	SizeBytes    int64             `json:"size_bytes"`
	URI          string            `json:"uri"`
	CreatedAt    time.Time         `json:"created_at"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// EpisodeRecord is the durable application state for one episode.
type EpisodeRecord struct {
	EpisodeID string            `json:"episode_id"`
	State     string            `json:"state"`
	UpdatedAt time.Time         `json:"updated_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Artifacts []ArtifactRef     `json:"artifacts,omitempty"`
}

// LocalStore is an offline filesystem implementation for tests and local MVP
// operation. It is intentionally interface-compatible with future Postgres/S3
// backends but does not require credentials or network access.
type LocalStore struct {
	root string
	now  func() time.Time
}

type artifactIndex struct {
	Artifacts map[string]ArtifactRef `json:"artifacts"`
}

// NewLocalStore creates a local durable store rooted inside a caller-provided
// directory.
func NewLocalStore(root string) (*LocalStore, error) {
	if strings.TrimSpace(root) == "" {
		return nil, fmt.Errorf("storage root is required")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0o700); err != nil {
		return nil, err
	}
	return &LocalStore{
		root: abs,
		now:  func() time.Time { return time.Now().UTC() },
	}, nil
}

// PutArtifact stores content by hash and links it to an immutable episode/name
// reference. Repeating the same write is idempotent; changing content for an
// existing name is rejected.
func (s *LocalStore) PutArtifact(ctx context.Context, input PutArtifactInput) (ArtifactRef, error) {
	if err := ctx.Err(); err != nil {
		return ArtifactRef{}, err
	}
	episodeID, err := safeSegment("episode_id", input.EpisodeID)
	if err != nil {
		return ArtifactRef{}, err
	}
	artifactName, err := safeSegment("artifact_name", input.ArtifactName)
	if err != nil {
		return ArtifactRef{}, err
	}
	if len(input.Content) == 0 {
		return ArtifactRef{}, fmt.Errorf("artifact content is required")
	}
	if input.ValidateCanonical {
		if err := s.validateCanonical(artifactName, input.Content); err != nil {
			return ArtifactRef{}, err
		}
	}

	contentHash := contentHash(input.Content)
	index, err := s.readIndex(episodeID)
	if err != nil {
		return ArtifactRef{}, err
	}
	if existing, ok := index.Artifacts[artifactName]; ok {
		if existing.ContentHash != contentHash {
			return ArtifactRef{}, fmt.Errorf("%w: %s already points to %s", ErrImmutableConflict, artifactName, existing.ContentHash)
		}
		return existing, nil
	}

	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = s.now()
	}
	ref := ArtifactRef{
		EpisodeID:    episodeID,
		ArtifactName: artifactName,
		ContentHash:  contentHash,
		SizeBytes:    int64(len(input.Content)),
		URI:          artifactURI(episodeID, artifactName, contentHash),
		CreatedAt:    createdAt.UTC(),
		Metadata:     copyMetadata(input.Metadata),
	}

	if err := s.writeObject(ref, input.Content); err != nil {
		return ArtifactRef{}, err
	}
	index.Artifacts[artifactName] = ref
	if err := s.writeIndex(episodeID, index); err != nil {
		return ArtifactRef{}, err
	}
	return ref, nil
}

// GetArtifact reads content for a previously returned reference and verifies
// the content hash before returning bytes.
func (s *LocalStore) GetArtifact(ctx context.Context, ref ArtifactRef) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	episodeID, err := safeSegment("episode_id", ref.EpisodeID)
	if err != nil {
		return nil, err
	}
	artifactName, err := safeSegment("artifact_name", ref.ArtifactName)
	if err != nil {
		return nil, err
	}
	index, err := s.readIndex(episodeID)
	if err != nil {
		return nil, err
	}
	stored, ok := index.Artifacts[artifactName]
	if !ok || (ref.ContentHash != "" && stored.ContentHash != ref.ContentHash) {
		return nil, ErrNotFound
	}
	data, err := os.ReadFile(s.objectPath(stored))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if got := contentHash(data); got != stored.ContentHash {
		return nil, fmt.Errorf("artifact hash mismatch: expected %s got %s", stored.ContentHash, got)
	}
	return data, nil
}

// ListArtifacts returns artifact refs for one episode in stable name order.
func (s *LocalStore) ListArtifacts(ctx context.Context, episodeID string) ([]ArtifactRef, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	safeEpisodeID, err := safeSegment("episode_id", episodeID)
	if err != nil {
		return nil, err
	}
	index, err := s.readIndex(safeEpisodeID)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(index.Artifacts))
	for name := range index.Artifacts {
		names = append(names, name)
	}
	sort.Strings(names)
	refs := make([]ArtifactRef, 0, len(names))
	for _, name := range names {
		refs = append(refs, index.Artifacts[name])
	}
	return refs, nil
}

// SaveEpisode stores one episode state record.
func (s *LocalStore) SaveEpisode(ctx context.Context, record EpisodeRecord) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	episodeID, err := safeSegment("episode_id", record.EpisodeID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(record.State) == "" {
		return fmt.Errorf("episode state is required")
	}
	if record.UpdatedAt.IsZero() {
		record.UpdatedAt = s.now()
	}
	record.EpisodeID = episodeID
	record.UpdatedAt = record.UpdatedAt.UTC()
	record.Metadata = copyMetadata(record.Metadata)
	record.Artifacts = append([]ArtifactRef(nil), record.Artifacts...)

	if err := os.MkdirAll(s.episodeDir(episodeID), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.statePath(episodeID), data, 0o600)
}

// GetEpisode loads one episode state record.
func (s *LocalStore) GetEpisode(ctx context.Context, episodeID string) (EpisodeRecord, error) {
	if err := ctx.Err(); err != nil {
		return EpisodeRecord{}, err
	}
	safeEpisodeID, err := safeSegment("episode_id", episodeID)
	if err != nil {
		return EpisodeRecord{}, err
	}
	data, err := os.ReadFile(s.statePath(safeEpisodeID))
	if err != nil {
		if os.IsNotExist(err) {
			return EpisodeRecord{}, ErrNotFound
		}
		return EpisodeRecord{}, err
	}
	var record EpisodeRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return EpisodeRecord{}, err
	}
	return record, nil
}

// AddArtifactRef links a stored artifact reference to an episode state record.
func (s *LocalStore) AddArtifactRef(ctx context.Context, episodeID string, ref ArtifactRef) error {
	if ref.EpisodeID != episodeID {
		return fmt.Errorf("artifact ref episode mismatch: %s != %s", ref.EpisodeID, episodeID)
	}
	record, err := s.GetEpisode(ctx, episodeID)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return err
		}
		record = EpisodeRecord{EpisodeID: episodeID, State: "backlog"}
	}
	for _, existing := range record.Artifacts {
		if existing.ArtifactName == ref.ArtifactName && existing.ContentHash == ref.ContentHash {
			return nil
		}
	}
	record.Artifacts = append(record.Artifacts, ref)
	record.UpdatedAt = s.now()
	return s.SaveEpisode(ctx, record)
}

func (s *LocalStore) validateCanonical(artifactName string, content []byte) error {
	stagingRoot := filepath.Join(s.root, ".validation")
	if err := os.MkdirAll(stagingRoot, 0o700); err != nil {
		return err
	}
	dir, err := os.MkdirTemp(stagingRoot, "artifact-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, artifactName)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		return err
	}
	report := artifacts.ValidatePath(path)
	if report.Valid {
		return nil
	}
	return fmt.Errorf("%w: %v", ErrValidationFailed, artifacts.ValidateReport(report))
}

func (s *LocalStore) readIndex(episodeID string) (artifactIndex, error) {
	index := artifactIndex{Artifacts: map[string]ArtifactRef{}}
	data, err := os.ReadFile(s.indexPath(episodeID))
	if err != nil {
		if os.IsNotExist(err) {
			return index, nil
		}
		return index, err
	}
	if err := json.Unmarshal(data, &index); err != nil {
		return index, err
	}
	if index.Artifacts == nil {
		index.Artifacts = map[string]ArtifactRef{}
	}
	return index, nil
}

func (s *LocalStore) writeIndex(episodeID string, index artifactIndex) error {
	if err := os.MkdirAll(s.episodeDir(episodeID), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.indexPath(episodeID), data, 0o600)
}

func (s *LocalStore) writeObject(ref ArtifactRef, content []byte) error {
	path := s.objectPath(ref)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.WriteFile(path, content, 0o600)
}

func (s *LocalStore) episodeDir(episodeID string) string {
	return filepath.Join(s.root, "episodes", episodeID)
}

func (s *LocalStore) indexPath(episodeID string) string {
	return filepath.Join(s.episodeDir(episodeID), "artifacts.json")
}

func (s *LocalStore) statePath(episodeID string) string {
	return filepath.Join(s.episodeDir(episodeID), "state.json")
}

func (s *LocalStore) objectPath(ref ArtifactRef) string {
	hash := strings.TrimPrefix(ref.ContentHash, "sha256:")
	prefix := hash
	if len(prefix) > 2 {
		prefix = prefix[:2]
	}
	return filepath.Join(s.root, "objects", prefix, hash, ref.ArtifactName)
}

func contentHash(content []byte) string {
	sum := sha256.Sum256(content)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func artifactURI(episodeID string, artifactName string, hash string) string {
	return "local://artifact-store/" + episodeID + "/" + artifactName + "@" + hash
}

func safeSegment(field string, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s is required", field)
	}
	if value == "." || value == ".." || strings.Contains(value, "..") {
		return "", fmt.Errorf("%s contains unsafe path traversal", field)
	}
	if strings.ContainsAny(value, `/\:`) {
		return "", fmt.Errorf("%s must be a single safe path segment", field)
	}
	return value, nil
}

func copyMetadata(metadata map[string]string) map[string]string {
	if len(metadata) == 0 {
		return nil
	}
	out := make(map[string]string, len(metadata))
	for key, value := range metadata {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key != "" && value != "" {
			out[key] = value
		}
	}
	return out
}
