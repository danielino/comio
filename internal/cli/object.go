package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

// objectCmd represents the object command
var objectCmd = &cobra.Command{
	Use:   "object",
	Short: "Object management commands",
}

var objectPutCmd = &cobra.Command{
	Use:   "put <bucket> <key> <file>",
	Short: "Put an object",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		bucket := args[0]
		key := args[1]
		filePath := args[2]

		file, err := os.Open(filePath)
		if err != nil {
			fmt.Printf("Error opening file: %v\n", err)
			return
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			fmt.Printf("Error getting file info: %v\n", err)
			return
		}
		fileSize := fileInfo.Size()

		// TODO: Get server address from config
		url := fmt.Sprintf("%s/%s/%s", serverAddr, bucket, key)

		req, err := http.NewRequest("PUT", url, file)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			return
		}

		req.ContentLength = fileSize

		// Set content type if possible, or let server guess
		// req.Header.Set("Content-Type", "application/octet-stream")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error sending request: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Error uploading object: %s\n", resp.Status)
			return
		}

		fmt.Printf("Successfully uploaded object %s/%s\n", bucket, key)
	},
}

var objectListCmd = &cobra.Command{
	Use:   "list <bucket> [prefix]",
	Short: "List objects in a bucket",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bucket := args[0]
		prefix := ""
		if len(args) > 1 {
			prefix = args[1]
		}

		url := fmt.Sprintf("%s/%s", serverAddr, bucket)
		if prefix != "" {
			url += "?prefix=" + prefix
		}

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
			fmt.Printf("Error listing objects: %s\n", resp.Status)
			os.Exit(1)
		}

		// Read and parse response
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			fmt.Printf("Error decoding response: %v\n", err)
			os.Exit(1)
		}

		objects, ok := result["Objects"].([]interface{})
		if !ok || len(objects) == 0 {
			fmt.Printf("No objects found in bucket %s\n", bucket)
			return
		}

		fmt.Printf("Objects in bucket %s:\n", bucket)
		for _, obj := range objects {
			o := obj.(map[string]interface{})
			key, keyOk := o["key"].(string)
			if !keyOk {
				key, keyOk = o["Key"].(string)
			}
			size, sizeOk := o["size"].(float64)
			if !sizeOk {
				size, _ = o["Size"].(float64)
			}
			if keyOk {
				fmt.Printf("  %s (%d bytes)\n", key, int64(size))
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(objectCmd)
	objectCmd.AddCommand(objectPutCmd)
	objectCmd.AddCommand(objectListCmd)
}
