package testutils

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"time"
)

type S3TestContainer struct {
	Container testcontainers.Container
	Endpoint  string
	Port      string
	Client    *s3.Client
}

// StartS3Container starts a MinIO container for S3 testing
func StartS3Container(ctx context.Context) (*S3TestContainer, error) {
	accessKey := "test-access-key"
	secretKey := "test-secret-key"

	req := testcontainers.ContainerRequest{
		Image: "minio/minio:latest",
		Env: map[string]string{
			"MINIO_ACCESS_KEY": accessKey,
			"MINIO_SECRET_KEY": secretKey,
		},
		ExposedPorts: []string{"9000/tcp"},
		Cmd:          []string{"server", "/data"},
		WaitingFor:   wait.ForListeningPort("9000/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:         true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start MinIO container: %v", err)
	}

	port, err := container.MappedPort(ctx, "9000")
	if err != nil {
		return nil, fmt.Errorf("failed to get container port: %v", err)
	}

	endpoint := fmt.Sprintf("http://localhost:%s", port.Port())

	// Create S3 client
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID {
			return aws.Endpoint{
				URL:           endpoint,
				SigningRegion: "us-east-1",
				Source:        aws.EndpointSourceCustom,
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	awsCfg := aws.Config{
		Credentials: credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		Region:      "us-east-1",
		EndpointResolverWithOptions: customResolver,
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	s3Container := &S3TestContainer{
		Container: container,
		Endpoint:  endpoint,
		Port:      port.Port(),
		Client:    client,
	}

	return s3Container, nil
}

// CreateBucket creates a new bucket in the test container
func (s *S3TestContainer) CreateBucket(ctx context.Context, bucket string) error {
	_, err := s.Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return fmt.Errorf("failed to create bucket: %v", err)
	}

	// Wait for the bucket to be available
	waiter := s3.NewBucketExistsWaiter(s.Client)
	err = waiter.Wait(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	}, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to wait for bucket: %v", err)
	}

	return nil
}

// GetTestS3Config returns a test S3 configuration
func (s *S3TestContainer) GetTestS3Config(bucket string) map[string]string {
	return map[string]string{
		"endpoint":        s.Endpoint,
		"region":          "us-east-1", // MinIO default region
		"bucket":          bucket,
		"accessKeyID":     "test-access-key",
		"secretAccessKey": "test-secret-key",
	}
}
