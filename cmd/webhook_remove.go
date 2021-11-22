package cmd

import (
	"context"
	"fmt"
	"github-admin-tool/restclient"
	"log"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

var webhookRemoveCmd = &cobra.Command{ // nolint // needed for cobra
	Use:   "webhook-remove",
	Short: "Remove webhook settings for repos in provided list by hostname",
	RunE:  webhookRemoveRun,
}

// nolint // needed for cobra
func init() {
	webhookRemoveCmd.Flags().StringVarP(
		&reposFile, "repos", "r", "", "path to file containing repositories (file should contain repos on new line without org/ prefix)",
	)
	webhookRemoveCmd.Flags().StringVarP(&webhookHost, "host", "n", "", "hostname to remove webhook for")
	webhookRemoveCmd.MarkFlagRequired("repos")
	webhookRemoveCmd.MarkFlagRequired("host")
	webhookRemoveCmd.Flags().SortFlags = true
	rootCmd.AddCommand(webhookRemoveCmd)
}

func webhookRemoveRun(cmd *cobra.Command, args []string) error {
	err := removeWebhookCommand(
		cmd,
		&repository{
			reader: &repositoryReaderService{},
		},
	)

	return err
}

func removeWebhookCommand(cmd *cobra.Command, repo *repository) error {
	host, reposFilePath, dryRun, err := removeWebhookFlagCheck(cmd)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	repositoryList, err := repo.reader.read(reposFilePath)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if dryRun {
		log.Printf("This is a dry run, the run would process %d repositories", len(repositoryList))

		return nil
	}

	ctx := context.Background()

	for _, repositoryName := range repositoryList {
		webhookID := getWebhookID(ctx, host, repositoryName)
		if webhookID > 0 {
			log.Printf("Removing %s for repo %s id is %d", host, repositoryName, webhookID)

			if err = removeWebhook(ctx, webhookID, repositoryName); err != nil {
				return fmt.Errorf("%w", err)
			}
		}
	}

	return nil
}

func removeWebhook(ctx context.Context, webhookID int, repositoryName string) error {
	client := restclient.NewClient(
		fmt.Sprintf("/repos/%s/%s/hooks/%d", config.Org, repositoryName, webhookID),
		config.Token,
		http.MethodDelete,
	)

	var response interface{}

	if err := client.Run(ctx, response); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func getWebhookID(ctx context.Context, host, repositoryName string) (webhookID int) {
	// Get webhooks and find ID if they match the host
	client := restclient.NewClient(
		fmt.Sprintf("/repos/%s/%s/hooks", config.Org, repositoryName),
		config.Token,
		http.MethodGet,
	)

	response := []WebhookResponse{}
	if err := client.Run(ctx, &response); err == nil {
		for _, webhook := range response {
			if webhook.Config.URL == host {
				return webhook.ID
			}
		}
	}

	log.Printf("No webhook for %s found in %s", host, repositoryName)

	return webhookID
}

func removeWebhookFlagCheck(cmd *cobra.Command) (host, reposFilePath string, dryRun bool, err error) {
	dryRun, err = cmd.Flags().GetBool("dry-run")
	if err != nil {
		return host, reposFilePath, dryRun, fmt.Errorf("%w", err)
	}

	host, err = cmd.Flags().GetString("host")
	if err != nil {
		return host, reposFilePath, dryRun, fmt.Errorf("%w", err)
	}

	_, err = url.ParseRequestURI(host)
	if err != nil {
		return host, reposFilePath, dryRun, fmt.Errorf("%w", err)
	}

	reposFilePath, err = cmd.Flags().GetString("repos")
	if err != nil {
		return host, reposFilePath, dryRun, fmt.Errorf("%w", err)
	}

	return host, reposFilePath, dryRun, nil
}
