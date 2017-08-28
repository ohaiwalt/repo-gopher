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

// func TestEnsureLabel(t *testing.T) {
// 	setup()
// 	defer teardown()

// 	// Get all labels
// 	mux.HandleFunc("/repos/org/r/labels", func(w http.ResponseWriter, r *http.Request) {
// 		testMethod(t, r, "GET")
// 		testFormValues(t, r, values{"page": "2"})
// 		fmt.Fprint(w, `[{"name": "a"},{"name": "b"}]`)
// 	})

// 	opt := &github.ListOptions{Page: 2}
// 	labels, _, err := client.Issues.ListLabels(context.Background(), "org", "r", opt)
// 	if err != nil {
// 		t.Errorf("Issues.ListLabels returned error: %v", err)
// 	}

// 	want := []*github.Label{{Name: github.String("a")}, {Name: github.String("b")}}
// 	if !reflect.DeepEqual(labels, want) {
// 		t.Errorf("Issues.ListLabels returned %+v, want %+v", labels, want)
// 	}

// 	// Get one label
// 	mux.HandleFunc("/repos/o/r/labels/n", func(w http.ResponseWriter, r *http.Request) {
// 		testMethod(t, r, "GET")
// 		fmt.Fprint(w, `{"url":"u", "name": "n", "color": "c"}`)
// 	})

// 	label, _, err := client.Issues.GetLabel(context.Background(), "o", "r", "n")
// 	if err != nil {
// 		t.Errorf("Issues.GetLabel returned error: %v", err)
// 	}

// 	want1 := &github.Label{URL: github.String("u"), Name: github.String("n"), Color: github.String("c")}
// 	if !reflect.DeepEqual(label, want) {
// 		t.Errorf("Issues.GetLabel returned %+v, want %+v", label, want1)
// 	}

// 	// Edit label
// 	input := &github.Label{Name: github.String("z")}

// 	mux.HandleFunc("/repos/org/r/labels/n", func(w http.ResponseWriter, r *http.Request) {
// 		v := new(Label)
// 		json.NewDecoder(r.Body).Decode(v)

// 		testMethod(t, r, "PATCH")
// 		if !reflect.DeepEqual(v, input) {
// 			t.Errorf("Request body = %+v, want %+v", v, input)
// 		}

// 		fmt.Fprint(w, `{"url":"u"}`)
// 	})

// 	label, _, err = client.Issues.EditLabel(context.Background(), "org", "r", "n", input)
// 	if err != nil {
// 		t.Errorf("Issues.EditLabel returned error: %v", err)
// 	}

// 	want2 := &github.Label{URL: github.String("u")}
// 	if !reflect.DeepEqual(label, want) {
// 		t.Errorf("Issues.EditLabel returned %+v, want %+v", label, want2)
// 	}

// 	// Create label
// 	input = &github.Label{Name: github.String("n")}

// 	mux.HandleFunc("/repos/o/r/labels", func(w http.ResponseWriter, r *http.Request) {
// 		v := new(Label)
// 		json.NewDecoder(r.Body).Decode(v)

// 		testMethod(t, r, "POST")
// 		if !reflect.DeepEqual(v, input) {
// 			t.Errorf("Request body = %+v, want %+v", v, input)
// 		}

// 		fmt.Fprint(w, `{"url":"u"}`)
// 	})

// 	label, _, err = client.Issues.CreateLabel(context.Background(), "o", "r", input)
// 	if err != nil {
// 		t.Errorf("Issues.CreateLabel returned error: %v", err)
// 	}

// 	want3 := &github.Label{URL: github.String("u")}
// 	if !reflect.DeepEqual(label, want) {
// 		t.Errorf("Issues.CreateLabel returned %+v, want %+v", label, want3)
// 	}

// 	// Delete label
// 	mux.HandleFunc("/repos/org/repo/labels/n", func(w http.ResponseWriter, r *http.Request) {
// 		testMethod(t, r, "DELETE")
// 	})

// 	_, err = client.Issues.DeleteLabel(context.Background(), "org", "repo", "n")
// 	if err != nil {
// 		t.Errorf("Issues.DeleteLabel returned error: %v", err)
// 	}

// }

func TestIssuesWith(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/search/issues", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"q":        "blah",
			"sort":     "forks",
			"order":    "desc",
			"page":     "2",
			"per_page": "2",
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

// func TestIsLabelInSlice(t *testing.T) {
// 	labels := []*github.Label{
// 		github.Label{Name: github.String("a")},
// 	}
// }

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
