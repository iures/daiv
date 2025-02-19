package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	internalGithub "daiv/internal/github"

	"github.com/google/go-github/v68/github"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RepositoryConfig holds the configuration for each repository.
type RepositoryConfig struct {
	Owner    string   `mapstructure:"owner"`
	Repo     string   `mapstructure:"repo"`
	Keywords []string `mapstructure:"keywords"`
}

// Config represents our overall configuration for the command.
type Config struct {
	Repositories []RepositoryConfig `mapstructure:"repositories"`
}

// getConfig extracts and validates the configuration for relevantPrs.
func getConfig() (Config, error) {
	sub := viper.Sub("relevantPrs")
	if sub == nil {
		return Config{}, fmt.Errorf("no configuration found for relevantPrs; please add a 'relevantPrs' section to your config file")
	}

	var cfg Config
	if err := sub.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("error unmarshaling relevantPrs config: %w", err)
	}

	if len(cfg.Repositories) == 0 {
		return Config{}, fmt.Errorf("no repositories configured")
	}

	return cfg, nil
}

// listAllPRs collects all pull requests with pagination support.
func listAllPRs(ctx context.Context, client *github.Client, owner, repo string, options *github.PullRequestListOptions) ([]*github.PullRequest, error) {
	var allPRs []*github.PullRequest
	opts := *options // create a copy so as not to modify the original options
	opts.ListOptions = github.ListOptions{PerPage: 100}

	for {
		prs, resp, err := client.PullRequests.List(ctx, owner, repo, &opts)
		if err != nil {
			return nil, err
		}
		allPRs = append(allPRs, prs...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allPRs, nil
}

// findKeywordMatches scans the diff text and returns any lines that match one or more keywords (case-insensitive)
func findKeywordMatches(diffStr string, keywords []string) []string {
	var matchedLines []string
	// Precompute lower-case keywords to avoid repetition.
	lowerKeywords := make([]string, len(keywords))
	for i, keyword := range keywords {
		lowerKeywords[i] = strings.ToLower(keyword)
	}

	for _, line := range strings.Split(diffStr, "\n") {
		lowerLine := strings.ToLower(line)
		for _, lowerKeyword := range lowerKeywords {
			if strings.Contains(lowerLine, lowerKeyword) {
				matchedLines = append(matchedLines, line)
				break
			}
		}
	}

	return matchedLines
}

// processRepository handles querying a single repository and outputting the matching PR changes.
func processRepository(ctx context.Context, client *github.Client, repoConfig RepositoryConfig) {
	prList, err := listAllPRs(ctx, client, repoConfig.Owner, repoConfig.Repo, &github.PullRequestListOptions{
		State: "open",
	})

	if err != nil {
		log.Printf("Error listing PRs for %s/%s: %v", repoConfig.Owner, repoConfig.Repo, err)
		return
	}

	for index, pr := range prList {
		diff, _, err := client.PullRequests.GetRaw(ctx, repoConfig.Owner, repoConfig.Repo, pr.GetNumber(), github.RawOptions{Type: github.Diff})

		if err != nil {
			log.Printf("Error getting diff for PR #%d: %v", pr.GetNumber(), err)
			continue
		}

		diffStr := string(diff)

		matchedLines := findKeywordMatches(diffStr, repoConfig.Keywords)

		if len(matchedLines) > 0 {
      if index == 0 {
        fmt.Printf("Repository: %s/%s\n", repoConfig.Owner, repoConfig.Repo)
      }

      fmt.Printf("  (PR #%d)[%s]: \n  %s\n", pr.GetNumber(), pr.GetHTMLURL(), pr.GetTitle())

			fmt.Println("    Matched changes:")
			for _, mLine := range matchedLines {
				fmt.Printf("      %s\n", mLine)
			}
		}
	}

	fmt.Println("")
}

// relevantPrs is the main function for the command, orchestrating configuration reading,
// GitHub client creation, and concurrent processing of the repositories.
func relevantPrs() {
	cfg, err := getConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	ctx := context.Background()
	client, err := internalGithub.NewGithubClient()
	if err != nil {
		log.Fatalf("Error creating github client: %v", err)
	}

	var wg sync.WaitGroup
	for _, repoConfig := range cfg.Repositories {
		wg.Add(1)
		// Capture the variable to avoid race conditions.
		repoCfg := repoConfig
		go func() {
			defer wg.Done()
			processRepository(ctx, client, repoCfg)
		}()
	}
	wg.Wait()
}

// relevantPrsCmd represents the updated relevantPrs command with improved descriptions.
var relevantPrsCmd = &cobra.Command{
	Use:   "relevantPrs",
	Short: "Search open PRs for changes matching specific keywords",
	Long: `Searches through all open pull requests in specified repositories
and displays changes containing user-defined keywords. This helps in quickly
identifying the relevant code changes among many open PRs.`,
	Run: func(cmd *cobra.Command, args []string) {
		relevantPrs()
	},
}

func init() {
	rootCmd.AddCommand(relevantPrsCmd)
}
