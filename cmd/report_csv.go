package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func GenerateCsv(ignoreArchived bool, allResults []ReportResponse) {
	parsed := parse(ignoreArchived, allResults)
	lines := writeCsv(parsed)
	err := writeToFile(lines)
	if err != nil {
		fmt.Println(err)
	}
}

func parse(ignoreArchived bool, allResults []ReportResponse) [][]string {
	var parsed [][]string
	for _, allData := range allResults {
		for _, repo := range allData.Organization.Repositories.Nodes {
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

func writeCsv(parsed [][]string) [][]string {
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

func writeToFile(lines [][]string) error {
	file, err := os.Create("report.csv")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	err = writer.WriteAll(lines)

	return err
}
