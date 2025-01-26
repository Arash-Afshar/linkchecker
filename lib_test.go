package linkchecker

import (
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDomainBlacklistWhitelist(t *testing.T) {
	tests := []struct {
		blacklist []string
		whitelist []string
		links     []string
		isValid   []bool
	}{
		{
			[]string{"a.com"},
			[]string{"b.com"},
			[]string{
				"https://a.com",
				"https://b.com",
				"https://c.com",
			},
			[]bool{false, false, false}, // since you cannot have both whitelist and blacklist
		},
		{
			[]string{},
			[]string{},
			[]string{
				"https://a.com",
				"https://b.com",
				"https://c.com",
			},
			[]bool{true, true, true}, // default to all all
		},
		{
			[]string{"a.com"},
			[]string{},
			[]string{
				"a.com",
				"https://a.com",
				"https://www.a.com",
				"https://www.a.com/test",
				"http://test.a.com/test",
				"https://b.com",
				"https://www.b.com",
				"https://www.b.com/test",
				"http://test.b.com/test",
			},
			[]bool{false, false, false, false, false, true, true, true, true},
		},
		{
			[]string{},
			[]string{"a.com"},
			[]string{
				"a.com",
				"https://a.com",
				"https://www.a.com",
				"https://www.a.com/test",
				"http://test.a.com/test",
				"https://b.com",
				"https://www.b.com",
				"https://www.b.com/test",
				"http://test.b.com/test",
			},
			[]bool{true, true, true, true, true, false, false, false, false},
		},
	}

	for testIndex, test := range tests {
		links := []Link{}
		for _, link := range test.links {
			links = append(links, Link{URL: link, IsValid: false})
		}
		validateLinks(links, test.whitelist, test.blacklist)
		for i, link := range links {
			if link.IsValid != test.isValid[i] {
				t.Errorf("Test %d: Link %d is not valid", testIndex, i)
			}
		}
	}
}

func TestLivenesscheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/dead" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(``))
		} else if r.URL.Path == "/access-denied" {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(``))
		} else if r.URL.Path == "/redirect" {
			w.Header().Set("Location", strings.Replace(r.URL.Path, "/redirect", "/live", 1))
			w.WriteHeader(http.StatusFound)
			w.Write([]byte(``))
		} else if r.URL.Path == "/live" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(``))
		} else {
			t.Errorf("Expected to request '/dead' or '/live', got: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	links := []Link{
		{URL: server.URL + "/dead"},
		{URL: server.URL + "/live"},
		{URL: server.URL + "/redirect"},
		{URL: server.URL + "/access-denied"},
	}

	livenesscheck(links)
	for _, link := range links {
		if link.URL == server.URL+"/live" && !link.IsLive {
			t.Errorf("Link %s is not live", link.URL)
		} else if link.URL == server.URL+"/redirect" && !link.IsLive {
			t.Errorf("Link %s is not live", link.URL)
		} else if link.URL == server.URL+"/dead" && link.IsLive {
			t.Errorf("Link %s is live", link.URL)
		} else if link.URL == server.URL+"/access-denied" && link.IsLive {
			t.Errorf("Link %s is live", link.URL)
		}
	}
}

func TestExtractLinks(t *testing.T) {
	links, err := extractLinks([]string{"test-data/test.pdf"})
	if err != nil {
		t.Errorf("Error extracting links: %v", err)
	}
	sort.Slice(links, func(i, j int) bool {
		return links[i].URL < links[j].URL
	})
	assert.Equal(t, len(links), 3)
	assert.Equal(t, links, []Link{
		{URL: "https://a.com", IsValid: false},
		{URL: "https://b.com", IsValid: false},
		{URL: "https://c.com", IsValid: false},
	})
}
