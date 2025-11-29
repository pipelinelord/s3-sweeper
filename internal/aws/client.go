package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pipelineload/s3-sweeper/internal/worker"
)

// S3Client wraps the AWS SDK client to provide a clean interface for our app
type S3Client struct {
	Client *s3.Client
}

// DeleteObject deletes a single object from S3.
func (s *S3Client) DeleteObject(ctx context.Context, bucket, key string) error {
	_, err := s.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	return err
}

// NewS3Client initializes a new AWS session using the default credential chain.
// This means it will look for env vars (AWS_ACCESS_KEY_ID) or ~/.aws/credentials automatically.
func NewS3Client(ctx context.Context, region string) (*S3Client, error) {
	// LoadDefaultConfig reads the configuration from the environment
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}

	// Create the S3 service client
	client := s3.NewFromConfig(cfg)

	return &S3Client{Client: client}, nil
}

// ListObjectsPages iterates over a bucket and pushes objects into the job channel.
// It closes the channel when iteration is complete.
func (s *S3Client) ListObjectsPages(ctx context.Context, bucket string, jobChan chan<- worker.Job) error {
	// 1. Always close the channel when we are done, or the workers will hang forever waiting for data.
	defer close(jobChan)

	// 2. Setup the input for the S3 List command
	input := &s3.ListObjectsV2Input{
		Bucket: &bucket,
	}

	// 3. Create the Paginator
	paginator := s3.NewListObjectsV2Paginator(s.Client, input)

	fmt.Println("Starting S3 Listing...")

	pageNum := 0
	for paginator.HasMorePages() {
		// Get the next page
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list objects: %w", err)
		}

		// Loop through objects in this page
		for _, object := range output.Contents {
			// Create the Job
			job := worker.Job{
				Key:          *object.Key,
				Bucket:       bucket,
				Size:         aws.ToInt64(object.Size),
				LastModified: *object.LastModified,
			}

			// PUSH to channel. This blocks if the channel is full (backpressure).
			select {
			case jobChan <- job:
				// Success
			case <-ctx.Done():
				// If user hits Ctrl+C, stop pushing
				return ctx.Err()
			}
		}
		pageNum++
		if pageNum%10 == 0 {
			log.Printf("Scanned %d pages...", pageNum)
		}
	}

	return nil
}
