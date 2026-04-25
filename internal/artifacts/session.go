package artifacts

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"harrierops-azure/internal/models"
)

var errInvalidSessionArtifact = errors.New("invalid session artifact")

type ExpectedSession struct {
	Command          string
	SchemaVersion    string
	ToolVersion      string
	TenantID         string
	SubscriptionID   string
	CurrentPrincipal models.ArtifactPrincipal
	AuthMode         string
	TokenSource      string
	CommandOptions   map[string]string
	MaxAge           time.Duration
	Now              time.Time
}

type SessionLoadResult[T any] struct {
	Payload T
	Source  models.SessionArtifact
}

type SessionAnchor struct {
	TenantID         string
	SubscriptionID   string
	CurrentPrincipal models.ArtifactPrincipal
	AuthMode         string
	TokenSource      string
}

type sessionMetadata struct {
	AuthMode        *string                 `json:"auth_mode"`
	Command         string                  `json:"command"`
	GeneratedAt     string                  `json:"generated_at"`
	SchemaVersion   string                  `json:"schema_version"`
	SubscriptionID  *string                 `json:"subscription_id"`
	TenantID        *string                 `json:"tenant_id"`
	TokenSource     *string                 `json:"token_source"`
	ArtifactContext *models.ArtifactContext `json:"artifact_context"`
}

func LoadSessionArtifact[T any](workspace string, expected ExpectedSession) (SessionLoadResult[T], bool, error) {
	var zero SessionLoadResult[T]
	if workspace == "" {
		workspace = "."
	}
	for _, path := range candidatePaths(workspace, expected.Command) {
		result, ok, err := loadCandidate[T](path, expected)
		if err != nil {
			if errors.Is(err, errInvalidSessionArtifact) {
				continue
			}
			return zero, false, err
		}
		if ok {
			return result, true, nil
		}
	}
	return zero, false, nil
}

func HasSessionArtifact(workspace string, command string) bool {
	if workspace == "" {
		workspace = "."
	}
	for _, path := range candidatePaths(workspace, command) {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}

func LoadSessionAnchor(workspace string, schemaVersion string, toolVersion string, maxAge time.Duration, now time.Time) (SessionAnchor, bool, error) {
	return LoadSessionAnchorFromCommands(workspace, []string{"whoami"}, schemaVersion, toolVersion, maxAge, now)
}

func LoadSessionAnchorFromCommands(workspace string, commands []string, schemaVersion string, toolVersion string, maxAge time.Duration, now time.Time) (SessionAnchor, bool, error) {
	var zero SessionAnchor
	if workspace == "" {
		workspace = "."
	}
	for _, command := range commands {
		for _, path := range candidatePaths(workspace, command) {
			_, metadata, ok, err := loadCandidateDataAndMetadata(path)
			if err != nil {
				if errors.Is(err, errInvalidSessionArtifact) {
					continue
				}
				return zero, false, err
			}
			if !ok {
				continue
			}
			if _, ok := validSessionAnchorMetadata(metadata, command, schemaVersion, toolVersion, maxAge, now); !ok {
				continue
			}
			return SessionAnchor{
				TenantID:         stringPtrValue(metadata.TenantID),
				SubscriptionID:   stringPtrValue(metadata.SubscriptionID),
				CurrentPrincipal: metadata.ArtifactContext.CurrentPrincipal,
				AuthMode:         stringPtrValue(metadata.AuthMode),
				TokenSource:      stringPtrValue(metadata.TokenSource),
			}, true, nil
		}
	}
	return zero, false, nil
}

func candidatePaths(workspace string, command string) []string {
	return []string{
		filepath.Join(workspace, "json", command+".json"),
		filepath.Join(workspace, "loot", command+".json"),
	}
}

func loadCandidate[T any](path string, expected ExpectedSession) (SessionLoadResult[T], bool, error) {
	var zero SessionLoadResult[T]
	data, metadata, ok, err := loadCandidateDataAndMetadata(path)
	if err != nil || !ok {
		return zero, false, err
	}
	generatedAt, ok := validSessionMetadata(metadata, expected)
	if !ok {
		return zero, false, nil
	}

	var payload T
	if err := json.Unmarshal(data, &payload); err != nil {
		return zero, false, fmt.Errorf("%w: read session artifact payload %s: %v", errInvalidSessionArtifact, path, err)
	}
	return SessionLoadResult[T]{
		Payload: payload,
		Source: models.SessionArtifact{
			Command:     expected.Command,
			Path:        path,
			GeneratedAt: generatedAt.Format(time.RFC3339),
			AgeSeconds:  int(expected.Now.Sub(generatedAt).Seconds()),
			Context:     "same tenant, subscription, principal, command options",
		},
	}, true, nil
}

func loadCandidateDataAndMetadata(path string) ([]byte, sessionMetadata, bool, error) {
	var zero sessionMetadata
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, zero, false, nil
	}
	if err != nil {
		return nil, zero, false, err
	}

	var header struct {
		Metadata sessionMetadata `json:"metadata"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return nil, zero, false, fmt.Errorf("%w: read session artifact metadata %s: %v", errInvalidSessionArtifact, path, err)
	}
	return data, header.Metadata, true, nil
}

func validSessionAnchorMetadata(metadata sessionMetadata, command string, schemaVersion string, toolVersion string, maxAge time.Duration, now time.Time) (time.Time, bool) {
	generatedAt, err := time.Parse(time.RFC3339, metadata.GeneratedAt)
	if err != nil {
		return time.Time{}, false
	}
	if now.IsZero() || now.Sub(generatedAt) < 0 || now.Sub(generatedAt) > maxAge {
		return time.Time{}, false
	}
	if metadata.Command != command ||
		metadata.SchemaVersion != schemaVersion ||
		stringPtrValue(metadata.TenantID) == "" ||
		stringPtrValue(metadata.SubscriptionID) == "" ||
		stringPtrValue(metadata.AuthMode) == "" ||
		stringPtrValue(metadata.TokenSource) == "" ||
		metadata.ArtifactContext == nil ||
		metadata.ArtifactContext.ToolVersion != toolVersion ||
		metadata.ArtifactContext.CurrentPrincipal.ID == "" ||
		metadata.ArtifactContext.CurrentPrincipal.TenantID == "" ||
		!reflect.DeepEqual(metadata.ArtifactContext.CommandOptions, map[string]string{}) {
		return time.Time{}, false
	}
	return generatedAt, true
}

func validSessionMetadata(metadata sessionMetadata, expected ExpectedSession) (time.Time, bool) {
	generatedAt, err := time.Parse(time.RFC3339, metadata.GeneratedAt)
	if err != nil {
		return time.Time{}, false
	}
	if expected.Now.IsZero() || expected.Now.Sub(generatedAt) < 0 || expected.Now.Sub(generatedAt) > expected.MaxAge {
		return time.Time{}, false
	}
	if metadata.Command != expected.Command ||
		metadata.SchemaVersion != expected.SchemaVersion ||
		stringPtrValue(metadata.TenantID) != expected.TenantID ||
		stringPtrValue(metadata.SubscriptionID) != expected.SubscriptionID ||
		stringPtrValue(metadata.AuthMode) != expected.AuthMode ||
		stringPtrValue(metadata.TokenSource) != expected.TokenSource ||
		metadata.ArtifactContext == nil ||
		metadata.ArtifactContext.ToolVersion != expected.ToolVersion ||
		metadata.ArtifactContext.CurrentPrincipal.ID != expected.CurrentPrincipal.ID ||
		metadata.ArtifactContext.CurrentPrincipal.TenantID != expected.CurrentPrincipal.TenantID ||
		!reflect.DeepEqual(metadata.ArtifactContext.CommandOptions, expected.CommandOptions) {
		return time.Time{}, false
	}
	return generatedAt, true
}

func stringPtrValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
