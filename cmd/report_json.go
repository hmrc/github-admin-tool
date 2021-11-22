package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

type reportJSON interface {
	generate(bool, []ReportResponse, map[string]string) ([]byte, error)
	generateWebhook(map[string][]WebhookResponse) ([]byte, error)
	uploader(string, []byte) error
}

type reportJSONService struct{}

func (r *reportJSONService) uploader(filePath string, reportJSON []byte) error {
	if err := ioutil.WriteFile(filePath, reportJSON, 0600); err != nil {
		return fmt.Errorf("failed to create %s: %w", filePath, err)
	}

	log.Printf("Report written to %s", filePath)

	return nil
}

func (r *reportJSONService) generate(
	ignoreArchived bool,
	allResults []ReportResponse,
	teamAccess map[string]string,
) ([]byte, error) {
	var repos []RepositoriesNode

	for _, allData := range allResults {
		for _, repo := range allData.Organization.Repositories.Nodes { // nolint // not modifying
			if ignoreArchived && repo.IsArchived {
				continue
			}

			repo.TeamPermissions = teamAccess[repo.Name]
			repos = append(repos, repo)
		}
	}

	reportJSON, err := json.Marshal(repos)

	if err != nil || len(repos) == 0 {
		return nil, fmt.Errorf("failed to marshal: %w", err)
	}

	return reportJSON, nil
}

func (r *reportJSONService) generateWebhook(allResults map[string][]WebhookResponse) ([]byte, error) {
	reportJSON, err := json.Marshal(allResults)

	if err != nil || len(allResults) == 0 {
		return nil, fmt.Errorf("failed to marshal: %w", err)
	}

	return reportJSON, nil
}
