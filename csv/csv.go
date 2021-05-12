package csv

import (
	"encoding/csv"
	"fmt"
	"os"
)

func Run(parsed [][]string) {
	f, e := os.Create("./repos.csv")
	if e != nil {
		fmt.Println(e)
	}

	writer := csv.NewWriter(f)
	var data = [][]string{
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
			"IsAdminEnforced",
			"RequiresCommitSignatures",
			"RestrictsPushes",
			"RequiresApprovingReviews",
			"RequiresStatusChecks",
			"RequiresCodeOwnerReviews",
			"DismissesStaleReviews",
			"RequiresStrictStatusChecks",
			"RequiredApprovingReviewCount",
			"AllowsForcePushes",
			"AllowsDeletions",
			"Branch Protection Pattern",
		},
	}
	for _, line := range parsed {
		data = append(data, line)
	}

	e = writer.WriteAll(data)
	if e != nil {
		fmt.Println(e)
	}
}
