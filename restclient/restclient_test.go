package restclient

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
		return t.returnValue, errors.New("fail") // nolint // only mock error for test
	}

	return t.returnValue, nil
}

func Test_bodyReaderService_read(t *testing.T) {
	var client = &http.Client{
		Timeout: time.Duration(1000) * time.Millisecond,
	}
	bodyErrorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1")
	}))
	mockResponse, err := client.Get(bodyErrorServer.URL)
	if err != nil {
		t.Errorf("bodyReaderService.read() could not setup read failure %+v", err)
		return
	}

	defer mockResponse.Body.Close()

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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.body = tt.stringToRead
			b := &bodyReaderService{}
			gotResult, err := b.read(tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("bodyReaderService.read() error = %+v, wantErr %+v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("bodyReaderService.read() = %v, want %v", gotResult, tt.wantResult)
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
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewClient(tt.args.path, tt.args.token); !reflect.DeepEqual(got, tt.want) {
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
	}

	type args struct {
		ctx  context.Context
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
			args: args{
				ctx: ctx,
			},
		},
		{
			name:    "Run fails on ascii in endpoint",
			wantErr: true,
			fields: fields{
				httpClient: http.DefaultClient,
				endpoint:   "▒",
			},
			args: args{
				ctx: ctx,
			},
		},
		{
			name:    "Run fails on status code and decode",
			wantErr: true,
			fields: fields{
				httpClient: http.DefaultClient,
				endpoint:   "https://api.github.com/rate_limit",
			},
			args: args{
				ctx: ctx,
			},
			mockHTTPStatusCode: 401,
			mockHTTPReturnFile: "../testdata/mockEmptyResponse.json",
		},
		{
			name:    "Run fails on status code",
			wantErr: true,
			fields: fields{
				httpClient: http.DefaultClient,
				endpoint:   "https://api.github.com/rate_limit",
			},
			args: args{
				ctx: ctx,
			},
			mockHTTPStatusCode: 404,
			mockHTTPReturnFile: "../testdata/mockRest404Response.json",
		},
		{
			name:    "Run fails on unauthorized code",
			wantErr: true,
			fields: fields{
				httpClient: http.DefaultClient,
				endpoint:   "https://api.github.com/rate_limit",
			},
			args: args{
				ctx: ctx,
			},
			mockHTTPStatusCode: 401,
			mockHTTPReturnFile: "../testdata/mockRest401Response.json",
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
			},
			args: args{
				ctx: ctx,
			},
			mockHTTPStatusCode: 200,
			mockHTTPReturnFile: "../testdata/mockRestEmptyBodyResponse.json",
		},
		{
			name:    "Run fails on unmarshalling",
			wantErr: true,
			fields: fields{
				httpClient: http.DefaultClient,
				endpoint:   "https://api.github.com/rate_limit",
			},
			args: args{
				ctx: ctx,
			},
			mockHTTPStatusCode: 200,
			mockHTTPReturnFile: "../testdata/mockEmptyResponse.json",
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
			},
			args: args{
				ctx: ctx,
			},
			mockHTTPStatusCode: 200,
			mockHTTPReturnFile: "../testdata/mockRestRateLimitResponse.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockHTTPReturnFile != "" {
				mockHTTPReturn, err := ioutil.ReadFile(tt.mockHTTPReturnFile)
				if err != nil {
					t.Fatalf("failed to read test data: %v", err)
				}

				httpmock.RegisterResponder(
					"GET",
					tt.fields.endpoint,
					httpmock.NewStringResponder(tt.mockHTTPStatusCode, string(mockHTTPReturn)),
				)
			}

			c := &Client{
				endpoint:   tt.fields.endpoint,
				token:      tt.fields.token,
				httpClient: tt.fields.httpClient,
				closeReq:   tt.fields.closeReq,
				bodyReader: &tt.fields.bodyReader,
			}
			if err := c.Run(tt.args.ctx, tt.args.resp); (err != nil) != tt.wantErr {
				t.Errorf("Client.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
