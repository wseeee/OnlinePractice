package model

import (
	"OnlinePrictice/internal/storage"
	"log"
	"os"
)

var Store storage.ObjectStorage

func InitMinIO() {
	backend := os.Getenv("STORAGE_BACKEND")
	if backend == "" {
		backend = "minio"
	}

	if backend == "local" {
		log.Println("STORAGE_BACKEND=local, skipping MinIO init")
		return
	}

	s, err := storage.NewMinIOStorage()
	if err != nil {
		log.Printf("Warning: MinIO init failed: %v (avatar/upload features disabled)", err)
		return
	}
	Store = s
	log.Println("MinIO storage initialized")
}
