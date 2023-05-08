package handlers

import (
	"encoding/json"
	"fmt"
)

type Query struct {
	Path string `json:"path"`
}

// query takes an HTTP payload from HQ, and proxies it to the Kubernetes
// API. Note that the permissions are based on RBAC permissions of the agent
func (h *Handler) query(rawQuery string) error {
	query := &Query{}
	err := json.Unmarshal([]byte(rawQuery), query)
	if err != nil {
		return err
	}

	fmt.Println("Querying data")
	return nil
}
