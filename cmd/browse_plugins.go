package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v68/github"
	"github.com/spf13/cobra"
)

var browsePluginsCmd = &cobra.Command{
	Use:   "browse",
	Short: "Browse daiv plugins in public repositories",
	Long: `Browse daiv plugins in public repositories.
Finds repositories that have the topic 'daiv-plugin' associated with them.

Example:
  daiv plugin browse
  daiv plugin browse --sort stars
  daiv plugin browse --limit 20
  daiv plugin browse --filter "git integration"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		// Create an unauthenticated client for public repositories
		client := github.NewClient(nil)
		
		// Get sort flag
		sortFlag, _ := cmd.Flags().GetString("sort")
		
		// Get order flag 
		orderFlag, _ := cmd.Flags().GetString("order")
		
		// Get limit flag
		limitFlag, _ := cmd.Flags().GetInt("limit")
		
		// Get filter flag
		filterFlag, _ := cmd.Flags().GetString("filter")
		
		// Search for repositories with the daiv-plugin topic
		searchQuery := "topic:daiv-plugin"
		
		// If filter is specified, search in description and readme too
		if filterFlag != "" {
			searchQuery += fmt.Sprintf(" %s in:description,readme", filterFlag)
		}
		
		searchOpts := &github.SearchOptions{
			Sort: sortFlag,
			Order: orderFlag,
			ListOptions: github.ListOptions{
				PerPage: limitFlag,
			},
		}
		
		fmt.Println("Browsing repositories with the daiv-plugin topic...\n")
		
		result, _, err := client.Search.Repositories(ctx, searchQuery, searchOpts)
		if err != nil {
			return fmt.Errorf("failed to search repositories: %w", err)
		}
		
		if len(result.Repositories) == 0 {
			fmt.Println("No daiv plugins found in public repositories.")
			return nil
		}
		
		fmt.Printf("Found %d daiv plugins:\n\n", len(result.Repositories))
		
		for _, repo := range result.Repositories {
			// Get plugin name (keep full name if it doesn't have daiv- prefix)
			pluginName := repo.GetName()
			if strings.HasPrefix(pluginName, "daiv-") {
				pluginName = strings.TrimPrefix(pluginName, "daiv-")
			}
			
			// Format the output
			fmt.Printf("- %s\n", pluginName)
			fmt.Printf("  Repository: %s\n", repo.GetFullName())
			
			// Print description if available
			if desc := repo.GetDescription(); desc != "" {
				fmt.Printf("  Description: %s\n", desc)
			}
			
			// Print stars and forks
			fmt.Printf("  Stars: %d | Forks: %d\n", repo.GetStargazersCount(), repo.GetForksCount())
			
			// Print last updated
			fmt.Printf("  Updated: %s\n", repo.GetUpdatedAt().Format("2006-01-02"))
			
			// Print installation command
			fmt.Printf("  Install: daiv plugin install %s\n", repo.GetFullName())
			
			fmt.Println()
		}
		
		return nil
	},
}

func init() {
	pluginCmd.AddCommand(browsePluginsCmd)
	
	// Add flags
	browsePluginsCmd.Flags().StringP("sort", "s", "updated", "Sort repositories by: stars, forks, updated")
	browsePluginsCmd.Flags().StringP("order", "o", "desc", "Order results: asc, desc")
	browsePluginsCmd.Flags().IntP("limit", "l", 10, "Limit the number of results")
	browsePluginsCmd.Flags().StringP("filter", "f", "", "Filter plugins by keyword in description or readme")
} 
