package csv

import (
	"encoding/csv"
	"fmt"
	"os"
)

func Run(parsed [][]string) {
	var lines = [][]string{
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
	for _, line := range parsed {
		lines = append(lines, line)
	}
	write(lines)
}

func write(lines [][]string) {
	f, e := os.Create("./report.csv")
	if e != nil {
		fmt.Println(e)
	}

	writer := csv.NewWriter(f)

	e = writer.WriteAll(lines)
	if e != nil {
		fmt.Println(e)
	}
}
