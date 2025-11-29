package worker

import (
	"context"
	"time"
)


// S3Deleter defines the behavior we need from the AWS client.
// This decouples the worker from the actual AWS implementation.
type S3Deleter interface {
	DeleteObject(ctx context.Context, bucket, key string) error
}

// Job represents a single S3 object that needs to be processed.
// We pass this through the "Jobs" channel.
type Job struct {
	Key          string    // The file path in S3
	Bucket       string    // The bucket name
	Size         int64     // File size in bytes
	LastModified time.Time // When it was last edited
}

// Result represents the outcome of processing a single Job.
// We pass this through the "Results" channel.
type Result struct {
	Key       string
	IsStale   bool  // True if the file met the "delete" criteria
	Size      int64 // How much space we (potentially) saved
	Err       error // If something went wrong checking this specific file
}