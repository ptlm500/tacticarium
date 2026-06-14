package seed

import (
	"encoding/json"
	"fmt"
	"os"
)

// gameVersion mirrors the {edition, dataslate} block present on every
// 40kdc-data entity.
type gameVersion struct {
	Edition   string `json:"edition"`
	Dataslate string `json:"dataslate"`
}

// readJSON reads filePath and unmarshals it into out.
func readJSON(filePath string, out any) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", filePath, err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("parsing %s: %w", filePath, err)
	}
	return nil
}
