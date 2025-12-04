package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

// bucketCmd represents the bucket command
var bucketCmd = &cobra.Command{
	Use:   "bucket",
	Short: "Bucket management commands",
}

var bucketCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new bucket",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bucketName := args[0]
		url := fmt.Sprintf("%s/%s", serverAddr, bucketName)

		req, err := http.NewRequest("PUT", url, nil)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			os.Exit(1)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error sending request: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("Error creating bucket: %s (Status: %d)\n", string(body), resp.StatusCode)
			os.Exit(1)
		}

		fmt.Printf("Bucket %s created successfully\n", bucketName)
	},
}

var bucketListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all buckets",
	Run: func(cmd *cobra.Command, args []string) {
		url := fmt.Sprintf("%s/", serverAddr)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			os.Exit(1)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error sending request: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("Error listing buckets: %s (Status: %d)\n", string(body), resp.StatusCode)
			os.Exit(1)
		}

		var buckets []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&buckets); err != nil {
			fmt.Printf("Error decoding response: %v\n", err)
			os.Exit(1)
		}

		for _, b := range buckets {
			fmt.Printf("%s\n", b["name"])
		}
	},
}

var bucketCountCmd = &cobra.Command{
	Use:   "count <name>",
	Short: "Count objects in a bucket",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bucket := args[0]

		url := fmt.Sprintf("%s/admin/%s/objects", serverAddr, bucket)

		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			os.Exit(1)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error sending request: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("Error getting bucket info: %s (Status: %d)\n", string(body), resp.StatusCode)
			os.Exit(1)
		}

		var info map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
			fmt.Printf("Error decoding response: %v\n", err)
			os.Exit(1)
		}

		count := int(info["count"].(float64))
		totalSize := int64(info["total_size"].(float64))

		fmt.Printf("Bucket '%s' contains %d object(s) (%s)\n", bucket, count, formatBytes(float64(totalSize)))
	},
}

func init() {
	rootCmd.AddCommand(bucketCmd)
	bucketCmd.AddCommand(bucketCreateCmd)
	bucketCmd.AddCommand(bucketListCmd)
	bucketCmd.AddCommand(bucketCountCmd)
}
