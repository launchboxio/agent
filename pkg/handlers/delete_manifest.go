package handlers

import "fmt"

func (h *Handler) deleteManifest(namespacedName string) error {
	fmt.Println("Deleting manifest...")
	fmt.Println(namespacedName)
	return nil
}
