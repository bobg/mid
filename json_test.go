package mid

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type jsonInput struct {
	A int    `json:"a"`
	B string `json:"b"`
}

type jsonOutput struct {
	C string `json:"c"`
	D int    `json:"d"`
}

func TestJSON(t *testing.T) {
	var received jsonInput

	mux := http.NewServeMux()

	// This is every combination of ctx/in/out/err being present or absent.
	mux.Handle("/a", JSON(func() {}))
	mux.Handle("/b", JSON(func(ctx context.Context) {}))
	mux.Handle("/c", JSON(func(in jsonInput) {
		received = in
	}))
	mux.Handle("/d", JSON(func() jsonOutput {
		return jsonOutput{C: "tock"}
	}))
	mux.Handle("/e", JSON(func() error {
		return CodeErr{C: http.StatusMethodNotAllowed}
	}))
	mux.Handle("/f", JSON(func(ctx context.Context, in jsonInput) {
		received = in
	}))
	mux.Handle("/g", JSON(func(context.Context) jsonOutput {
		return jsonOutput{D: 77}
	}))
	mux.Handle("/h", JSON(func(context.Context) error {
		return CodeErr{C: http.StatusPaymentRequired}
	}))
	mux.Handle("/i", JSON(func(in jsonInput) jsonOutput {
		return jsonOutput{C: in.B, D: in.A}
	}))
	mux.Handle("/j", JSON(func(in jsonInput) error {
		return CodeErr{C: in.A}
	}))
	mux.Handle("/k", JSON(func() (jsonOutput, error) {
		return jsonOutput{C: "reason"}, nil
	}))
	mux.Handle("/l", JSON(func(_ context.Context, in jsonInput) jsonOutput {
		return jsonOutput{C: in.B + in.B, D: 2 * in.A}
	}))
	mux.Handle("/m", JSON(func(ctx context.Context, in jsonInput) error {
		received = in
		return nil
	}))
	mux.Handle("/n", JSON(func(ctx context.Context) (jsonOutput, error) {
		return jsonOutput{}, CodeErr{C: http.StatusConflict}
	}))
	mux.Handle("/o", JSON(func(in jsonInput) (jsonOutput, error) {
		return jsonOutput{C: in.B[1:], D: in.A - 1}, nil
	}))
	mux.Handle("/p", JSON(func(ctx context.Context, in jsonInput) (jsonOutput, error) {
		return jsonOutput{C: strconv.Itoa(in.A), D: len(in.B)}, nil
	}))

	s := httptest.NewServer(mux)

	cases := []struct {
		path         string
		inp          string
		wantCode     int
		wantReceived *jsonInput
		wantOut      *jsonOutput
		ctx          context.Context
	}{
		{
			path:     "/a",
			wantCode: http.StatusNoContent,
		},
		{
			path:     "/b",
			wantCode: http.StatusNoContent,
		},
		{
			path:         "/c",
			inp:          `{"a":7, "b":"milo"}`,
			wantCode:     http.StatusNoContent,
			wantReceived: &jsonInput{A: 7, B: "milo"},
		},
		{
			path:     "/d",
			wantCode: http.StatusOK,
			wantOut:  &jsonOutput{C: "tock"},
		},
		{
			path:     "/e",
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			path:         "/f",
			inp:          `{"a":12, "b":"humbug"}`,
			wantCode:     http.StatusNoContent,
			wantReceived: &jsonInput{A: 12, B: "humbug"},
		},
		{
			path:     "/g",
			wantCode: http.StatusOK,
			wantOut:  &jsonOutput{D: 77},
		},
		{
			path:     "/h",
			wantCode: http.StatusPaymentRequired,
		},
		{
			path:     "/i",
			inp:      `{"a": 137, "b": "rhyme"}`,
			wantCode: http.StatusOK,
			wantOut:  &jsonOutput{C: "rhyme", D: 137},
		},
		{
			path:     "/j",
			inp:      `{"a": 599}`,
			wantCode: 599,
		},
		{
			path:     "/k",
			wantCode: http.StatusOK,
			wantOut:  &jsonOutput{C: "reason"},
		},
		{
			path:     "/l",
			inp:      `{"a": 206, "b": "azaz"}`,
			wantCode: http.StatusOK,
			wantOut:  &jsonOutput{C: "azazazaz", D: 412},
		},
		{
			path:         "/m",
			inp:          `{"b": "mathemagician"}`,
			wantCode:     http.StatusNoContent,
			wantReceived: &jsonInput{B: "mathemagician"},
		},
		{
			path:     "/n",
			wantCode: http.StatusConflict,
		},
		{
			path:     "/o",
			inp:      `{"a": 1234, "b": "5678"}`,
			wantCode: http.StatusOK,
			wantOut:  &jsonOutput{C: "678", D: 1233},
		},
		{
			path:     "/p",
			inp:      `{"a": 66, "b": "trivium"}`,
			wantCode: http.StatusOK,
			wantOut:  &jsonOutput{C: "66", D: 7},
		},
	}

	for i, c := range cases {
		var client http.Client

		t.Run(fmt.Sprintf("case_%02d", i+1), func(t *testing.T) {
			var inp io.Reader
			if c.inp != "" {
				inp = strings.NewReader(c.inp)
			}

			req, err := http.NewRequest("POST", s.URL+c.path, inp)
			if err != nil {
				t.Fatal(err)
			}

			if inp != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			received = jsonInput{}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != c.wantCode {
				t.Errorf("got code %d, want %d", resp.StatusCode, c.wantCode)
			}
			if c.wantOut != nil {
				var got jsonOutput
				err = json.NewDecoder(resp.Body).Decode(&got)
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(&got, c.wantOut); diff != "" {
					t.Errorf("out mismatch (-want +got):\n%s", diff)
				}
			}
			if c.wantReceived != nil {
				if diff := cmp.Diff(c.wantReceived, &received); diff != "" {
					t.Errorf("received mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
