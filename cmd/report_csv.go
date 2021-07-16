package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

var doReportCSVFileWrite = reportCSVFile // nolint // Like this for testing mock

func reportCSVGenerate(filePath string, ignoreArchived bool, allResults []ReportResponse) error {
	parsed := reportCSVParse(ignoreArchived, allResults)
	lines := reportCSVLines(parsed)

	if err := doReportCSVFileWrite(filePath, lines); err != nil {
		return fmt.Errorf("GenerateCSV failed: %w", err)
	}

	return nil
}

func reportCSVParse(ignoreArchived bool, allResults []ReportResponse) [][]string {
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
			"Parent Repo Name",
			"Merge Commit Allowed",
			"Squash Merge Allowed",
			"Rebase Merge Allowed",
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

func reportCSVFile(filePath string, lines [][]string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	if err = writer.WriteAll(lines); err != nil {
		return fmt.Errorf("failed to create %s: %w", filePath, err)
	}

	log.Printf("Report written to %s", filePath)

	return nil
}
