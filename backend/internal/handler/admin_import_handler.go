package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/peter/tacticarium/backend/internal/seed"
)

// Import handlers stay as raw chi handlers because they use multipart file uploads
// which are simpler to handle with http.Request directly.

func (h *AdminHandler) ImportFactions(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file upload required", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	tmpFile, err := writeTempFile(file, "factions-*.csv")
	if err != nil {
		http.Error(w, "failed to process upload", http.StatusInternalServerError)
		return
	}
	defer func() { _ = os.Remove(tmpFile) }()

	count, err := seed.SeedFactions(r.Context(), h.db, tmpFile)
	if err != nil {
		slog.ErrorContext(r.Context(), "Import factions error", "error", err)
		http.Error(w, "import failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"imported": count,
		"entity":   "factions",
	})
}

func (h *AdminHandler) ImportDetachments(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file upload required", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	tmpFile, err := writeTempFile(file, "detachments-*.csv")
	if err != nil {
		http.Error(w, "failed to process upload", http.StatusInternalServerError)
		return
	}
	defer func() { _ = os.Remove(tmpFile) }()

	count, err := seed.SeedDetachments(r.Context(), h.db, tmpFile)
	if err != nil {
		slog.ErrorContext(r.Context(), "Import detachments error", "error", err)
		http.Error(w, "import failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"imported": count,
		"entity":   "detachments",
	})
}

func (h *AdminHandler) ImportStratagems(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file upload required", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	tmpFile, err := writeTempFile(file, "stratagems-*.csv")
	if err != nil {
		http.Error(w, "failed to process upload", http.StatusInternalServerError)
		return
	}
	defer func() { _ = os.Remove(tmpFile) }()

	stratagems, err := seed.SeedStratagems(r.Context(), h.db, tmpFile)
	if err != nil {
		slog.ErrorContext(r.Context(), "Import stratagems error", "error", err)
		http.Error(w, "import failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"stratagems": stratagems,
		"entity":     "stratagems",
	})
}

func (h *AdminHandler) ImportMissions(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file upload required", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	tmpFile, err := writeTempFile(file, "missions-*.json")
	if err != nil {
		http.Error(w, "failed to process upload", http.StatusInternalServerError)
		return
	}
	defer func() { _ = os.Remove(tmpFile) }()

	stats, err := seed.SeedMissions(r.Context(), h.db, tmpFile)
	if err != nil {
		slog.ErrorContext(r.Context(), "Import missions error", "error", err)
		http.Error(w, "import failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"missions":        stats.Missions,
		"missionRules":    stats.MissionRules,
		"secondaries":     stats.Secondaries,
		"challengerCards": stats.ChallengerCards,
		"gambits":         stats.Gambits,
		"entity":          "missions",
	})
}

func writeTempFile(src interface{ Read([]byte) (int, error) }, pattern string) (string, error) {
	tmp, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}

	buf := make([]byte, 32*1024)
	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			if _, writeErr := tmp.Write(buf[:n]); writeErr != nil {
				_ = tmp.Close()
				_ = os.Remove(tmp.Name())
				return "", fmt.Errorf("writing temp file: %w", writeErr)
			}
		}
		if readErr != nil {
			break
		}
	}

	_ = tmp.Close()
	return filepath.Clean(tmp.Name()), nil
}
