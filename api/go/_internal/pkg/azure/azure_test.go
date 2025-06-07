package azure

import (
	"context"
	"log"
	"strings"
	"testing"

	appfx "github/huangc28/kikichoice-be/api/go/_internal/fx"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
)

type AzureTestSuite struct {
	suite.Suite
	client *BlobStorageWrapperClient
}

func (s *AzureTestSuite) SetupSuite() {
	fx.New(
		appfx.CoreConfigOptions,
		fx.Provide(
			NewSharedKeyCredential,
			NewBlobStorageClient,
			NewBlobStorageWrapperClient,
		),
		fx.Invoke(func(client *BlobStorageWrapperClient) {
			s.client = client
		}),
	)
}

func (s *AzureTestSuite) TestUploadTextFileToAzureBlob() {
	fileName := "test.txt"
	fileContent := "Hello, Azure Blob Storage! This is a test file."
	contentReader := strings.NewReader(fileContent)

	ctx := context.Background()

	url, err := s.client.UploadProductImage(ctx, fileName, contentReader)
	require.NoError(s.T(), err, "Failed to upload file to Azure blob storage")

	fileName2 := "subfolder/test2.txt"
	fileContent2 := "Hello, Azure Blob Storage! This is a test file 2."
	contentReader2 := strings.NewReader(fileContent2)

	url2, err := s.client.UploadProductImage(ctx, fileName2, contentReader2)
	require.NoError(s.T(), err, "Failed to upload file to Azure blob storage")

	log.Println("url", url)
	log.Println("url2", url2)
}

func TestAzureTestSuite(t *testing.T) {
	suite.Run(t, new(AzureTestSuite))
}
