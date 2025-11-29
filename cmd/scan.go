package cmd

import (
	"context"
	"fmt"

	"github.com/pipelineload/s3-sweeper/internal/aws"
	"github.com/pipelineload/s3-sweeper/internal/worker"
	"github.com/pipelineload/s3-sweeper/utils"
	"github.com/spf13/cobra"
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan a bucket for stale objects",
	Long: `Scans the specified S3 bucket using a concurrent worker pool.
Example: s3-sweeper scan --bucket my-app-logs --days 30 --workers 50`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Get Flags
		bucket, _ := cmd.Flags().GetString("bucket")
		days, _ := cmd.Flags().GetInt("days")
		workers, _ := cmd.Flags().GetInt("workers")
		region, _ := cmd.Flags().GetString("region")

		ctx := context.Background() // Create a base context

		fmt.Printf("Initializing S3 Sweeper...\n")
		fmt.Printf("Target: %s [%s]\n", bucket, region)

		shouldDelete, _ := cmd.Flags().GetBool("delete")

		if shouldDelete {
			fmt.Println("⚠️  WARNING: DELETE MODE ENABLED. Stale files will be removed!")
		} else {
			fmt.Println("ℹ️  DRY RUN: No files will be deleted.")
		}

		// 2. Initialize AWS Client
		s3Client, err := aws.NewS3Client(ctx, region)
		if err != nil {
			fmt.Printf("Error initializing AWS client: %v\n", err)
			return
		}

		fmt.Println("AWS Connection Established.")

		// Verification: Just print the pointer to ensure it's not nil
		fmt.Printf("Client ready: %v\n", s3Client)

		// 3. Setup Channels
		// Buffer size 100 means the AWS lister can get 100 items ahead of the workers
		// before it has to pause. This smooths out performance spikes.
		jobChan := make(chan worker.Job, 100)

		// 4. Start the Producer (S3 Lister) in a separate Goroutine
		// We MUST run this in a goroutine, otherwise it will block here forever
		// and the workers will never start.
		go func() {
			err := s3Client.ListObjectsPages(ctx, bucket, jobChan)
			if err != nil {
				fmt.Printf("Error listing objects: %v\n", err)
			}
		}()

		// 5. Start the Worker Pool
		resultsChan := worker.StartWorkerPool(s3Client, workers, days, !shouldDelete, jobChan)

		// 6. The Collector (Main Thread)
		// We read from resultsChan until it closes.
		var totalSize int64
		var count int
		var staleCount int
		var staleSize int64
		var deletedCount int

		fmt.Println("Scanning...")

		for result := range resultsChan {
			count++
			totalSize += result.Size

			if result.IsStale {
				staleCount++
				staleSize += result.Size

				// If we are in DELETE mode and there was no error, count it as deleted
				if shouldDelete && result.Err == nil {
					deletedCount++
					// Optional: Verbose logging
					// fmt.Printf("🗑️  Deleted: %s\n", result.Key)
				} else if result.Err != nil {
					fmt.Printf("❌ Error processing %s: %v\n", result.Key, result.Err)
				}
			}
		}

		// 7. Final Report
		fmt.Println("\n--- Scan Complete ---")
		fmt.Printf("Total Objects Scanned: %d\n", count)
		fmt.Printf("Stale Objects Found:   %d\n", staleCount)

		if shouldDelete {
			fmt.Printf("Objects Deleted:       %d\n", deletedCount)
		}

		fmt.Printf("Potential Space Savings: %s\n", utils.FormatBytes(staleSize))
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	// Local Flags for the scan command
	scanCmd.Flags().StringP("bucket", "b", "", "Name of the S3 bucket (required)")
	scanCmd.MarkFlagRequired("bucket")
	scanCmd.Flags().IntP("days", "d", 30, "Age in days to consider an object 'stale'")
	scanCmd.Flags().IntP("workers", "w", 10, "Number of concurrent workers")
	scanCmd.Flags().Bool("delete", false, "Perform actual deletion of stale objects")
}
