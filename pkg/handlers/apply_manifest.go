package handlers

import "fmt"

func (h *Handler) applyManifest(manifest string) error {
	fmt.Println("Applying manifest...")
	fmt.Println(manifest)
	return nil
}
