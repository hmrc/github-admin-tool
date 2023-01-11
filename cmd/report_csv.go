package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type reportCSV interface {
	opener(string) (*os.File, error)
	writer(*os.File, [][]string) error
}

type reportCSVService struct{}

func (r *reportCSVService) opener(filePath string) (file *os.File, err error) {
	file, err = os.Create(filePath)
	if err != nil {
		return file, fmt.Errorf("failed to create file: %w", err)
	}

	return file, nil
}

func (r *reportCSVService) writer(file *os.File, lines [][]string) error {
	defer file.Close()

	writer := csv.NewWriter(file)

	if err := writer.WriteAll(lines); err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filePath, err)
	}

	log.Printf("Report written to %s", filePath)

	return nil
}

func reportCSVUpload(service reportCSV, filePath string, lines [][]string) error {
	file, err := service.opener(filePath)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := service.writer(file, lines); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func reportCSVGenerate(ignoreArchived bool, allResults []ReportResponse, teamAccess map[string]string) [][]string {
	parsed := reportCSVParse(ignoreArchived, allResults, teamAccess)
	lines := reportCSVLines(parsed)

	return lines
}

func reportCSVWebhookGenerate(webhooks []Webhooks) [][]string {
	parsed := reportCSVWebhookParse(webhooks)
	lines := reportCSVWebhookLines(parsed)

	return lines
}

func reportCSVParse(ignoreArchived bool, allResults []ReportResponse, teamAccess map[string]string) [][]string {
	var parsed [][]string

	for _, allData := range allResults {
		for _, repo := range allData.Organization.Repositories.Nodes { // nolint // not modifying
			if ignoreArchived && repo.IsArchived {
				continue
			}

			repoSlice := []string{
				strings.TrimSpace(repo.NameWithOwner),
				strings.TrimSpace(repo.DefaultBranchRef.Name),
				strconv.FormatBool(repo.IsArchived),
				strconv.FormatBool(repo.IsPrivate),
				strconv.FormatBool(repo.IsEmpty),
				strconv.FormatBool(repo.IsFork),
				strings.TrimSpace(repo.Parent.NameWithOwner),
				strconv.FormatBool(repo.MergeCommitAllowed),
				strconv.FormatBool(repo.SquashMergeAllowed),
				strconv.FormatBool(repo.RebaseMergeAllowed),
				strings.TrimSpace(teamAccess[repo.Name]),
			}

			for _, protection := range repo.BranchProtectionRules.Nodes {
				repoSlice = append(repoSlice,
					strconv.FormatBool(protection.IsAdminEnforced),
					strconv.FormatBool(protection.RequiresCommitSignatures),
					strconv.FormatBool(protection.RestrictsPushes),
					strconv.FormatBool(protection.RequiresApprovingReviews),
					strconv.FormatBool(protection.RequiresStatusChecks),
					strconv.FormatBool(protection.RequiresCodeOwnerReviews),
					strconv.FormatBool(protection.DismissesStaleReviews),
					strconv.FormatBool(protection.RequiresStrictStatusChecks),
					strconv.Itoa(protection.RequiredApprovingReviewCount),
					strconv.FormatBool(protection.AllowsForcePushes),
					strconv.FormatBool(protection.AllowsDeletions),
					strings.TrimSpace(protection.Pattern),
				)
			}

			parsed = append(parsed, repoSlice)
		}
	}

	return parsed
}

func reportCSVLines(parsed [][]string) [][]string {
	lines := [][]string{
		{
			"Repo Name",
			"Default Branch Name",
			"Is Archived",
			"Is Private",
			"Is Empty",
			"Is Fork",
			"Has Wiki Enabled",
			"Parent Repo Name",
			"Merge Commit Allowed",
			"Squash Merge Allowed",
			"Rebase Merge Allowed",
			"Team Permissions",
			"(BP1) IsAdminEnforced",
			"(BP1) RequiresCommitSignatures",
			"(BP1) RestrictsPushes",
			"(BP1) RequiresApprovingReviews",
			"(BP1) RequiresStatusChecks",
			"(BP1) RequiresCodeOwnerReviews",
			"(BP1) DismissesStaleReviews",
			"(BP1) RequiresStrictStatusChecks",
			"(BP1) RequiredApprovingReviewCount",
			"(BP1) AllowsForcePushes",
			"(BP1) AllowsDeletions",
			"(BP1) Branch Protection Pattern",
			"(BP2) IsAdminEnforced",
			"(BP2) RequiresCommitSignatures",
			"(BP2) RestrictsPushes",
			"(BP2) RequiresApprovingReviews",
			"(BP2) RequiresStatusChecks",
			"(BP2) RequiresCodeOwnerReviews",
			"(BP2) DismissesStaleReviews",
			"(BP2) RequiresStrictStatusChecks",
			"(BP2) RequiredApprovingReviewCount",
			"(BP2) AllowsForcePushes",
			"(BP2) AllowsDeletions",
			"(BP2) Branch Protection Pattern",
		},
	}
	lines = append(lines, parsed...)

	return lines
}

func reportCSVWebhookParse(allResults []Webhooks) [][]string {
	var parsed [][]string

	for _, webhooks := range allResults {
		for _, webhook := range webhooks.Webhooks {
			repoSlice := []string{
				strings.TrimSpace(webhooks.RepositoryName),
				strconv.Itoa(webhook.ID),
				strings.TrimSpace(webhook.Config.URL),
				strconv.FormatBool(webhook.Active),
				strconv.Itoa(webhook.Config.InsecureURL),
				fmt.Sprintf("%+v", webhook.Events),
			}

			parsed = append(parsed, repoSlice)
		}
	}

	return parsed
}

func reportCSVWebhookLines(parsed [][]string) [][]string {
	lines := [][]string{
		{
			"Repo Name",
			"Webhook ID",
			"Webhook URL",
			"Is Active",
			"Insecure URL",
			"Events",
		},
	}
	lines = append(lines, parsed...)

	return lines
}
