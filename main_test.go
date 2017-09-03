package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/google/go-github/github"
)

var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// client is the GitHub client being tested.
	client *github.Client

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server
)

func TestIssuesWith(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/search/issues", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"q": "repo:o/r label:\"blah\"",
		})

		fmt.Fprint(w, `{"total_count": 4, "incomplete_results": true, "items": [{"number":1},{"number":2}]}`)
	})

	result, err := issuesWith(context.Background(), "o", "r", "label", "blah", client)
	if err != nil {
		t.Errorf("issuesWith returned error: %v", err)
	}

	want := &github.IssuesSearchResult{
		Issues: []github.Issue{{Number: github.Int(1)}, {Number: github.Int(2)}},
	}
	if !reflect.DeepEqual(result, want.Issues) {
		t.Errorf("issuesWith returned %+v, want %+v", result, want.Issues)
	}

}

func TestIsLabelInSlice(t *testing.T) {
	labels := []*github.Label{
		&github.Label{Name: github.String("a")},
		&github.Label{Name: github.String("b")},
	}

	result := isLabelInSlice("a", labels)
	want := true

	if !reflect.DeepEqual(result, want) {
		t.Errorf("issuesWith returned %+v, want %+v", result, want)
	}
}

// setup sets up a test HTTP server along with a github.Client that is
// configured to talk to that test server. Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup() {
	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	// github client configured to use test server
	client = github.NewClient(nil)
	url, _ := url.Parse(server.URL + "/")
	client.BaseURL = url
	client.UploadURL = url
}

// teardown closes the test HTTP server.
func teardown() {
	server.Close()
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

type values map[string]string

func testFormValues(t *testing.T, r *http.Request, values values) {
	want := url.Values{}
	for k, v := range values {
		want.Set(k, v)
	}

	r.ParseForm()
	if got := r.Form; !reflect.DeepEqual(got, want) {
		t.Errorf("Request parameters: %v, want %v", got, want)
	}
}
