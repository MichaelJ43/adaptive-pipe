package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	addr := ":8082"
	if v := os.Getenv("LISTEN_ADDR"); v != "" {
		addr = v
	}
	dataDir := os.Getenv("FILE_DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	nodeID := os.Getenv("FILE_NODE_ID")
	if nodeID == "" {
		nodeID = "file-local-1"
	}

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	http.HandleFunc("/v1/workspace/init", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			TenantSlug string `json:"tenant_slug"`
			GithubOrg  string `json:"github_org"`
			GithubRepo string `json:"github_repo"`
			RunID      string `json:"run_id"`
			CommitSHA  string `json:"commit_sha"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "json", http.StatusBadRequest)
			return
		}
		dir := filepath.Join(dataDir, req.TenantSlug, sanit(req.GithubOrg), sanit(req.GithubRepo), sanit(req.RunID))
		if err := os.MkdirAll(dir, 0750); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		meta := filepath.Join(dir, "HEAD")
		if err := os.WriteFile(meta, []byte(req.CommitSHA), 0600); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"node_id": nodeID,
			"path":    dir,
		})
	})

	log.Printf("file service %s node=%s data=%s", addr, nodeID, dataDir)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func sanit(s string) string {
	if s == "" {
		return "_"
	}
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r == '/' || r == '\\' || r == '.' {
			out = append(out, '_')
			continue
		}
		out = append(out, r)
	}
	return string(out)
}
