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
	mockRateLimitResponseFile      = "../testdata/mockRestRateLimitResponse.json"
	mockRateLimitEmptyResponseFile = "../testdata/mockRestRateLimitEmptyResponse.json"
	mockRestEmptyBodyResponseFile  = "../testdata/mockRestEmptyBodyResponse.json"
)

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
	fail bool
}

func (m *mockReportCSV) opener(filePath string) (file *os.File, err error) {
	if m.fail {
		return nil, errTestFail
	}

	return nil, nil
}

func (m *mockReportCSV) writer(file *os.File, lines [][]string) error {
	if m.fail {
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

func (m *mockReportJSON) generateWebhook(map[string][]WebhookResponse) ([]byte, error) {
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

func mockJsonMarshalError(v interface{}) ([]byte, error) {
	return []byte{}, errors.New("Marshalling failed")
}

type mockTeportWebhookGetter interface {
	getRepositoryList(*reportWebhook) ([]repositoryCursorList, error)
	getWebhooks(*reportWebhook, []repositoryCursorList) (map[string][]WebhookResponse, error)
}

type mockReportWebhookGetterService struct {
	failRepoList               bool
	failWebhook                bool
	returnRepoList             []repositoryCursorList
	returnRepositoryCursorList map[string][]WebhookResponse
}

func (r *mockReportWebhookGetterService) getRepositoryList(report *reportWebhook) ([]repositoryCursorList, error) {
	if r.failRepoList {
		return r.returnRepoList, errors.New("fail")
	}

	return r.returnRepoList, nil
}

func (r *mockReportWebhookGetterService) getWebhooks(report *reportWebhook, list []repositoryCursorList) (map[string][]WebhookResponse, error) {
	if r.failWebhook {
		return r.returnRepositoryCursorList, errors.New("fail")
	}

	return r.returnRepositoryCursorList, nil
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
