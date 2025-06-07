package azure

import (
	"fmt"
	"log"

	"github/huangc28/kikichoice-be/api/go/_internal/configs"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

func NewSharedKeyCredential(cfg *configs.Config) (*azblob.SharedKeyCredential, error) {
	cred, err := azblob.NewSharedKeyCredential(cfg.Azure.BlobStorageAccountName, cfg.Azure.BlobStorageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create shared key credential: %v", err)
	}
	return cred, nil
}

func NewBlobStorageClient(cfg *configs.Config, cred *azblob.SharedKeyCredential) (*azblob.Client, error) {
	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", cfg.Azure.BlobStorageAccountName)

	log.Printf("serviceURL: %+v", cfg.Azure)
	log.Println("serviceURL", serviceURL)

	client, err := azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	return client, nil
}
