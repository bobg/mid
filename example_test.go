package mid

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
)

func ExampleJSON() {
	type (
		inType struct {
			Referer   bool `json:"referer"`
			UserAgent bool `json:"user_agent"`
		}
		outType struct {
			Referer   string `json:"referer"`
			UserAgent string `json:"user_agent"`
		}
	)

	handler := func(ctx context.Context, in inType) (outType, error) {
		var (
			req = Request(ctx)
			out outType
		)

		if in.Referer {
			out.Referer = req.Referer()
		}
		if in.UserAgent {
			out.UserAgent = req.UserAgent()
		}

		return out, nil
	}

	s := httptest.NewServer(JSON(handler))
	defer s.Close()

	inp := `{"user_agent":true}`

	req, err := http.NewRequest("POST", s.URL, strings.NewReader(inp))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ExampleJSON-agent/1.0")

	var c http.Client
	resp, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var out outType
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Server got user agent %s\n", out.UserAgent)

	// Output: Server got user agent ExampleJSON-agent/1.0
}
