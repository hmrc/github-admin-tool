package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
)

const (
	fmReadWrite        fs.FileMode = 0o600
	reportJSONUploader             = "failed to create %s: %w"
)

type reportJSON interface {
	generate(ignoreArchived bool, allResults []ReportResponse, teamAccess map[string]string) ([]byte, error)
	generateWebhook(allResults []Webhooks) ([]byte, error)
	uploader(filePath string, reportJSON []byte) error
}

type reportJSONService struct{}

func (r *reportJSONService) uploader(filePath string, reportJSON []byte) error {
	if err := os.WriteFile(filePath, reportJSON, fmReadWrite); err != nil {
		return fmt.Errorf(reportJSONUploader, filePath, err)
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
		for _, repo := range allData.Organization.Repositories.Nodes {
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

func (r *reportJSONService) generateWebhook(allResults []Webhooks) ([]byte, error) {
	reportJSON, err := json.Marshal(allResults)

	if err != nil || len(allResults) == 0 {
		return nil, fmt.Errorf("failed to marshal: %w", err)
	}

	return reportJSON, nil
}
