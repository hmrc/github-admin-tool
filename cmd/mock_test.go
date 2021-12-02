package cmd

import (
	"errors"
	"fmt"
	"github-admin-tool/graphqlclient"
	"io/ioutil"
	"log"
	"os"

	"github.com/jarcoal/httpmock"
)

var (
	errTestFail                    = errors.New("fail")
	errTestAccessFail              = errors.New("access fail")
	errTestMarshalFail             = errors.New("Marshalling failed")               // nolint // expected global
	mockRateLimitResponseFile      = "testdata/mockRestRateLimitResponse.json"      // nolint // expected global
	mockRateLimitEmptyResponseFile = "testdata/mockRestRateLimitEmptyResponse.json" // nolint // expected global
	mockRestEmptyBodyResponseFile  = "testdata/mockRestEmptyBodyResponse.json"      // nolint // expected global
	mockEmptyCSVReportRows         = [][]string{                                    // nolint // expected global
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
)

const MockRepoName = "some-org"

type mockReportGetter struct {
	fail bool
}

func (m *mockReportGetter) getReport() ([]ReportResponse, error) {
	if m.fail {
		return []ReportResponse{}, errTestFail
	}

	return []ReportResponse{}, nil
}

type mockReportCSV struct {
	failOpen  bool
	failWrite bool
}

func (m *mockReportCSV) opener(filePath string) (file *os.File, err error) {
	if m.failOpen {
		return nil, errTestFail
	}

	return nil, nil
}

func (m *mockReportCSV) writer(file *os.File, lines [][]string) error {
	if m.failWrite {
		return errTestFail
	}

	return nil
}

type mockReportJSON struct {
	failupload   bool
	failgenerate bool
}

func (m *mockReportJSON) uploader(filePath string, reportJSON []byte) error {
	if m.failupload {
		return errTestFail
	}

	return nil
}

func (m *mockReportJSON) generate(
	ignoreArchived bool,
	allResults []ReportResponse,
	teamAccess map[string]string,
) ([]byte, error) {
	if m.failgenerate {
		return nil, errTestFail
	}

	return nil, nil
}

func (m *mockReportJSON) generateWebhook([]Webhooks) ([]byte, error) {
	if m.failgenerate {
		return nil, errTestFail
	}

	return nil, nil
}

type mockReportAccess struct {
	fail        bool
	returnValue map[string]string
}

func (m *mockReportAccess) getReport() (map[string]string, error) {
	if m.fail {
		return m.returnValue, errTestAccessFail
	}

	return m.returnValue, nil
}

type mockRepositoryReader struct {
	readFail    bool
	returnValue []string
}

func (t *mockRepositoryReader) read(reposFile string) ([]string, error) {
	if t.readFail {
		return t.returnValue, errTestFail
	}

	return t.returnValue, nil
}

type mockRepositoryGetter struct {
	getFail     bool
	returnValue map[string]*RepositoriesNode
}

func (t *mockRepositoryGetter) get(
	repositoryList []string,
	sender *githubRepositorySender,
) (
	map[string]*RepositoriesNode,
	error,
) {
	if t.getFail {
		return t.returnValue, errTestFail
	}

	return t.returnValue, nil
}

type mockRepositorySender struct {
	sendFail    bool
	returnValue map[string]*RepositoriesNode
}

func (t *mockRepositorySender) send(req *graphqlclient.Request) (map[string]*RepositoriesNode, error) {
	if t.sendFail {
		return nil, errors.New("fail") // nolint // only mock error for test
	}

	if len(t.returnValue) > 0 {
		return t.returnValue, nil
	}

	return make(map[string]*RepositoriesNode), nil
}

type mockSender struct {
	sendFail bool
	action   string
}

func (t *mockSender) send(req *graphqlclient.Request) error {
	if t.sendFail {
		return errors.New(fmt.Sprintf("%s: test", t.action)) // nolint // only mock error for test
	}

	return nil
}

func mockJSONMarshalError(v interface{}) ([]byte, error) {
	return []byte{}, errTestMarshalFail
}

type mockReportWebhookGetterService struct {
	failRepoList      bool
	failWebhook       bool
	returnRepoList    []repositoryCursorList
	returnWebhookList []Webhooks
}

func (r *mockReportWebhookGetterService) getRepositoryList(report *reportWebhook) ([]repositoryCursorList, error) {
	if r.failRepoList {
		return r.returnRepoList, errTestFail
	}

	return r.returnRepoList, nil
}

func (r *mockReportWebhookGetterService) getWebhooks(
	report *reportWebhook,
	list []repositoryCursorList,
) ([]Webhooks, error) {
	if r.failWebhook {
		return r.returnWebhookList, errTestFail
	}

	return r.returnWebhookList, nil
}

func mockHTTPResponder(method, url, responseFile string, statusCode int) {
	response, err := ioutil.ReadFile(responseFile)
	if err != nil {
		log.Fatalf("failed to read test data: %v", err)
	}

	httpmock.RegisterResponder(
		method,
		url,
		httpmock.NewStringResponder(statusCode, string(response)),
	)
}
