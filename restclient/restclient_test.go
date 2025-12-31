package restclient

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
)

type mockBodyReader struct {
	readFail    bool
	returnValue []byte
}

func (t *mockBodyReader) read(body io.Reader) ([]byte, error) {
	if t.readFail {
		return t.returnValue, errors.New("fail") //nolint // only mock error for test
	}

	return t.returnValue, nil
}

func Test_bodyReaderService_read(t *testing.T) {
	client := &http.Client{
		Timeout: time.Duration(1000) * time.Millisecond,
	}
	bodyErrorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1")
	}))

	mockResponse, err := client.Get(bodyErrorServer.URL) //nolint // noctx lint fail - this is only for test so ignore
	if err != nil {
		t.Errorf("bodyReaderService.read() could not setup read failure %+v", err)

		return
	}

	defer func() {
		_ = mockResponse.Body.Close()
	}()

	type args struct {
		body io.Reader
	}

	tests := []struct {
		name         string
		b            *bodyReaderService
		args         args
		stringToRead io.Reader
		wantResult   []byte
		wantErr      bool
	}{
		{
			name:         "bodyReaderService_read fails",
			stringToRead: mockResponse.Body,
			wantErr:      true,
			wantResult:   []byte{},
		},
		{
			name:         "bodyReaderService_read success",
			stringToRead: strings.NewReader("{}"),
			wantResult:   []byte{123, 125},
			wantErr:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.args.body = test.stringToRead
			bodyReaderService := &bodyReaderService{}
			gotResult, err := bodyReaderService.read(test.args.body)

			if (err != nil) != test.wantErr {
				t.Errorf("bodyReaderService.read() error = %+v, wantErr %+v", err, test.wantErr)

				return
			}

			if !reflect.DeepEqual(gotResult, test.wantResult) {
				t.Errorf("bodyReaderService.read() = %v, want %v", gotResult, test.wantResult)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	type args struct {
		path  string
		token string
	}

	tests := []struct {
		name string
		args args
		want *Client
	}{
		{
			name: "NewClient success",
			args: args{
				path:  "/rate_limit",
				token: "TOKEN",
			},
			want: &Client{
				endpoint:   "https://api.github.com/rate_limit",
				token:      "TOKEN",
				httpClient: http.DefaultClient,
				closeReq:   true,
				bodyReader: &bodyReaderService{},
				method:     "GET",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewClient(tt.args.path, tt.args.token, http.MethodGet); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClient() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestClient_Run(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	type fields struct {
		endpoint   string
		token      string
		httpClient *http.Client
		closeReq   bool
		bodyReader mockBodyReader
		method     string
	}

	type args struct {
		resp interface{}
	}

	ctx := context.Background()

	tests := []struct {
		name               string
		fields             fields
		args               args
		mockHTTPReturnFile string
		mockHTTPStatusCode int
		wantErr            bool
	}{
		{
			name:    "Run fails on incorrect endpoint",
			wantErr: true,
			fields: fields{
				httpClient: http.DefaultClient,
				endpoint:   ":7878",
			},
			args: args{},
		},
		{
			name:    "Run fails on ascii in endpoint",
			wantErr: true,
			fields: fields{
				httpClient: http.DefaultClient,
				endpoint:   "▒",
			},
			args: args{},
		},
		{
			name:    "Run fails on status code and decode",
			wantErr: true,
			fields: fields{
				httpClient: http.DefaultClient,
				endpoint:   "https://api.github.com/rate_limit",
			},
			args:               args{},
			mockHTTPStatusCode: 401,
			mockHTTPReturnFile: "testdata/mockEmptyResponse.json",
		},
		{
			name:    "Run fails on status code",
			wantErr: true,
			fields: fields{
				httpClient: http.DefaultClient,
				endpoint:   "https://api.github.com/rate_limit",
			},
			args:               args{},
			mockHTTPStatusCode: 404,
			mockHTTPReturnFile: "testdata/mockRest404Response.json",
		},
		{
			name:    "Run fails on unauthorized code",
			wantErr: true,
			fields: fields{
				httpClient: http.DefaultClient,
				endpoint:   "https://api.github.com/rate_limit",
				method:     "GET",
			},
			args:               args{},
			mockHTTPStatusCode: 401,
			mockHTTPReturnFile: "testdata/mockRest401Response.json",
		},
		{
			name:    "Run fails on reading body",
			wantErr: true,
			fields: fields{
				httpClient: http.DefaultClient,
				endpoint:   "https://api.github.com/rate_limit",
				bodyReader: mockBodyReader{
					readFail: true,
				},
				method: "GET",
			},
			args:               args{},
			mockHTTPStatusCode: 200,
			mockHTTPReturnFile: "testdata/mockRestEmptyBodyResponse.json",
		},
		{
			name:    "Run fails on unmarshalling",
			wantErr: true,
			fields: fields{
				httpClient: http.DefaultClient,
				endpoint:   "https://api.github.com/rate_limit",
				method:     "GET",
			},
			args:               args{},
			mockHTTPStatusCode: 200,
			mockHTTPReturnFile: "testdata/mockEmptyResponse.json",
		},
		{
			name:    "Run is success",
			wantErr: false,
			fields: fields{
				httpClient: http.DefaultClient,
				endpoint:   "https://api.github.com/rate_limit",
				bodyReader: mockBodyReader{
					returnValue: []byte{123, 10, 125},
				},
				method: "GET",
			},
			args:               args{},
			mockHTTPStatusCode: 200,
			mockHTTPReturnFile: "testdata/mockRestRateLimitResponse.json",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.mockHTTPReturnFile != "" {
				mockHTTPReturn, err := os.ReadFile(test.mockHTTPReturnFile)
				if err != nil {
					t.Fatalf("failed to read test data: %v", err)
				}

				httpmock.RegisterResponder(
					test.fields.method,
					test.fields.endpoint,
					httpmock.NewStringResponder(test.mockHTTPStatusCode, string(mockHTTPReturn)),
				)
			}

			client := &Client{
				endpoint:   test.fields.endpoint,
				token:      test.fields.token,
				httpClient: test.fields.httpClient,
				closeReq:   test.fields.closeReq,
				bodyReader: &test.fields.bodyReader,
				method:     test.fields.method,
			}
			if err := client.Run(ctx, test.args.resp); (err != nil) != test.wantErr {
				t.Errorf("Client.Run() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
