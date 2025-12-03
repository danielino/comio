package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

// adminCmd represents the admin command
var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Administrative commands",
}

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Show server metrics",
	Run: func(cmd *cobra.Command, args []string) {
		url := fmt.Sprintf("%s/admin/metrics", serverAddr)

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
			fmt.Printf("Error getting metrics: %s (Status: %d)\n", string(body), resp.StatusCode)
			os.Exit(1)
		}

		var metrics map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
			fmt.Printf("Error decoding response: %v\n", err)
			os.Exit(1)
		}

		storage := metrics["storage"].(map[string]interface{})
		totalBytes := storage["TotalBytes"].(float64)
		usedBytes := storage["UsedBytes"].(float64)
		freeBytes := storage["FreeBytes"].(float64)
		
		fmt.Printf("Storage Metrics:\n")
		fmt.Printf("  Total: %s\n", formatBytes(totalBytes))
		fmt.Printf("  Used:  %s\n", formatBytes(usedBytes))
		fmt.Printf("  Free:  %s\n", formatBytes(freeBytes))
	},
}

var purgeCmd = &cobra.Command{
	Use:   "purge <bucket>",
	Short: "Delete all objects in a bucket",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bucket := args[0]
		
		// First, get info about what will be deleted
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
		
		if count == 0 {
			fmt.Printf("No objects to delete in bucket '%s'\n", bucket)
			return
		}
		
		fmt.Printf("\nWARNING: This will delete %d object(s) totaling %s from bucket '%s'\n", 
			count, formatBytes(float64(totalSize)), bucket)
		fmt.Print("Are you sure you want to proceed? (yes/no): ")
		
		var confirmation string
		fmt.Scanln(&confirmation)
		
		if confirmation != "yes" {
			fmt.Println("Operation cancelled")
			os.Exit(0)
		}
		
		// Perform actual deletion
		deleteURL := fmt.Sprintf("%s/admin/%s/objects?confirm=true", serverAddr, bucket)
		deleteReq, err := http.NewRequest("DELETE", deleteURL, nil)
		if err != nil {
			fmt.Printf("Error creating delete request: %v\n", err)
			os.Exit(1)
		}
		
		deleteResp, err := client.Do(deleteReq)
		if err != nil {
			fmt.Printf("Error performing deletion: %v\n", err)
			os.Exit(1)
		}
		defer deleteResp.Body.Close()
		
		if deleteResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(deleteResp.Body)
			fmt.Printf("Error deleting objects: %s (Status: %d)\n", string(body), deleteResp.StatusCode)
			os.Exit(1)
		}
		
		var deleteResult map[string]interface{}
		if err := json.NewDecoder(deleteResp.Body).Decode(&deleteResult); err != nil {
			fmt.Printf("Error decoding delete response: %v\n", err)
			os.Exit(1)
		}
		
		deletedCount := int(deleteResult["deleted_count"].(float64))
		freedSize := int64(deleteResult["freed_size"].(float64))
		
		fmt.Printf("\n✓ Deleted %d object(s), freed %s\n", deletedCount, formatBytes(float64(freedSize)))
	},
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes float64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%.0f B", bytes)
	}
	div, exp := float64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.2f %s", bytes/div, units[exp])
}

func init() {
	rootCmd.AddCommand(adminCmd)
	adminCmd.AddCommand(metricsCmd)
	adminCmd.AddCommand(purgeCmd)
}
