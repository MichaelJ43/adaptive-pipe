package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ValidateClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewValidateClient(base string) *ValidateClient {
	return &ValidateClient{
		BaseURL: base,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *ValidateClient) Validate(ctx context.Context, org, repo, commit string) error {
	body, _ := json.Marshal(map[string]string{
		"github_org": org, "github_repo": repo, "commit_sha": commit,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/v1/validate", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("validate: status %d: %s", res.StatusCode, string(b))
	}
	return nil
}

type FileClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewFileClient(base string) *FileClient {
	return &FileClient{
		BaseURL: base,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type WorkspaceInitResponse struct {
	NodeID string `json:"node_id"`
	Path   string `json:"path"`
}

func (c *FileClient) InitWorkspace(ctx context.Context, tenantSlug, org, repo, runID, commit string) (*WorkspaceInitResponse, error) {
	body, _ := json.Marshal(map[string]string{
		"tenant_slug": tenantSlug,
		"github_org":  org,
		"github_repo": repo,
		"run_id":      runID,
		"commit_sha":  commit,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/v1/workspace/init", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("file init: status %d: %s", res.StatusCode, string(b))
	}
	var out WorkspaceInitResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
