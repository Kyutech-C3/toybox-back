//go:build integration

package testutil

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/simesaba80/toybox-back/internal/infrastructure/config"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	localstackImage = "localstack/localstack:4.13.1"
	s3Port          = "4566/tcp"

	testS3Bucket = "toybox-test-bucket"
	testS3Dir    = "test-assets"
	testS3Region = "us-east-1"
)

var (
	s3Once          sync.Once
	s3Client        *s3.Client
	s3Container     testcontainers.Container
	s3Endpoint      string
	s3InitErr       error
	accessKey       = "test"
	secretAccessKey = "test"
)

// SetupTestS3 は LocalStack を使用したS3互換コンテナを立ち上げ、テスト用の s3.Client を返却します。
func SetupTestS3(tb testing.TB) *s3.Client {
	tb.Helper()

	s3Once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		var endpoint string
		s3Container, endpoint, s3InitErr = startLocalstackContainer(ctx)
		if s3InitErr != nil {
			return
		}
		s3Endpoint = endpoint

		s3Client, s3InitErr = newS3Client(ctx, endpoint)
		if s3InitErr != nil {
			return
		}

		if err := ensureBucket(ctx, s3Client, testS3Bucket); err != nil {
			s3InitErr = err
			return
		}

		// テストで参照される設定値を環境変数およびconfigパッケージに反映
		os.Setenv("S3_BUCKET", testS3Bucket)
		os.Setenv("S3_DIR", testS3Dir)
		os.Setenv("S3_BASE_URL", endpoint)
		os.Setenv("REGION_NAME", testS3Region)
		os.Setenv("ACCESS_KEY", accessKey)
		os.Setenv("SECRET_ACCESS_KEY", secretAccessKey)

		config.S3_BUCKET = testS3Bucket
		config.S3_DIR = testS3Dir
		config.S3_BASE_URL = endpoint
		config.REGION_NAME = testS3Region
	})

	if s3InitErr != nil {
		tb.Fatalf("テストS3初期化に失敗しました: %v", s3InitErr)
	}

	return s3Client
}

func startLocalstackContainer(ctx context.Context) (testcontainers.Container, string, error) {
	req := testcontainers.ContainerRequest{
		Image:        localstackImage,
		ExposedPorts: []string{s3Port},
		Env: map[string]string{
			"SERVICES":               "s3",
			"REGION_NAME":            testS3Region,
			"AWS_DEFAULT_REGION":     testS3Region,
			"AWS_ACCESS_KEY_ID":      accessKey,
			"AWS_SECRET_ACCESS_KEY":  secretAccessKey,
			"DATA_DIR":               "/tmp/localstack/data",
			"DEBUG":                  "1",
			"LOCALSTACK_HOST":        "localhost",
			"DEFAULT_REGION":         testS3Region,
			"HOSTNAME":               "localhost",
			"HOSTNAME_EXTERNAL":      "localhost",
			"EDGE_PORT":              "4566",
			"SERVICES_ENDPOINT_PORT": "4566",
		},
		WaitingFor: wait.ForListeningPort(s3Port).WithStartupTimeout(2 * time.Minute),
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to start localstack container: %w", err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch localstack host: %w", err)
	}

	mappedPort, err := c.MappedPort(ctx, s3Port)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch localstack port: %w", err)
	}

	endpoint := fmt.Sprintf("http://%s:%s", host, mappedPort.Port())
	return c, endpoint, nil
}

func newS3Client(ctx context.Context, endpoint string) (*s3.Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(testS3Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretAccessKey, "")),
		awsconfig.WithBaseEndpoint(endpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	}), nil
}

func ensureBucket(ctx context.Context, client *s3.Client, bucket string) error {
	_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		var alreadyOwned *s3types.BucketAlreadyOwnedByYou
		if errors.As(err, &alreadyOwned) {
			return nil
		}
		var alreadyExists *s3types.BucketAlreadyExists
		if errors.As(err, &alreadyExists) {
			return nil
		}
		return fmt.Errorf("failed to create test bucket: %w", err)
	}
	return nil
}

// TeardownS3 は LocalStack コンテナおよび関連リソースを解放します。
func TeardownS3() {
	if s3Client != nil {
		s3Client = nil
	}

	if s3Container != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = s3Container.Terminate(ctx)
		s3Container = nil
	}

	s3Once = sync.Once{}
	s3InitErr = nil
	s3Endpoint = ""
}
