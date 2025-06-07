package azure

import (
	"context"
	"strings"
	"testing"

	appfx "github/huangc28/kikichoice-be/api/go/_internal/fx"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
)

type AzureTestSuite struct {
	suite.Suite
	client *azblob.Client
}

func (s *AzureTestSuite) SetupSuite() {
	fx.New(
		appfx.CoreConfigOptions,
		fx.Provide(
			NewSharedKeyCredential,
			NewBlobStorageClient,
		),
		fx.Invoke(func(client *azblob.Client) {
			s.client = client
		}),
	)
}

func (s *AzureTestSuite) TestUploadTextFileToAzureBlob() {
	fileName := "test.txt"
	fileContent := "Hello, Azure Blob Storage! This is a test file."
	contentReader := strings.NewReader(fileContent)

	ctx := context.Background()

	_, err := s.client.UploadStream(ctx, "products", fileName, contentReader, nil)
	require.NoError(s.T(), err, "Failed to upload file to Azure blob storage")

	fileName2 := "subfolder/test2.txt"
	fileContent2 := "Hello, Azure Blob Storage! This is a test file 2."
	contentReader2 := strings.NewReader(fileContent2)

	_, err = s.client.UploadStream(ctx, "products", fileName2, contentReader2, nil)
	require.NoError(s.T(), err, "Failed to upload file to Azure blob storage")
}

func TestAzureTestSuite(t *testing.T) {
	suite.Run(t, new(AzureTestSuite))
}
