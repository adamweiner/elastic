// Copyright 2012 Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

package elastic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// CountService is a convenient service for determining the
// number of documents in an index. Use SearchService with
// a SearchType of count for counting with queries etc.
type CountService struct {
	client  *Client
	indices []string
	debug   bool
	pretty  bool
}

// CountResult is the result returned from using the Count API
// (http://www.elasticsearch.org/guide/reference/api/count/)
type CountResult struct {
	Count  int64      `json:"count"`
	Shards shardsInfo `json:"_shards,omitempty"`
}

func NewCountService(client *Client) *CountService {
	builder := &CountService{
		client: client,
		debug:  false,
		pretty: false,
	}
	return builder
}

func (s *CountService) Index(index string) *CountService {
	if s.indices == nil {
		s.indices = make([]string, 0)
	}
	s.indices = append(s.indices, index)
	return s
}

func (s *CountService) Indices(indices ...string) *CountService {
	if s.indices == nil {
		s.indices = make([]string, 0)
	}
	s.indices = append(s.indices, indices...)
	return s
}

func (s *CountService) Pretty(pretty bool) *CountService {
	s.pretty = pretty
	return s
}

func (s *CountService) Debug(debug bool) *CountService {
	s.debug = debug
	return s
}

func (s *CountService) Do() (int64, error) {
	// Build url
	urls := "/"

	// Indices part
	indexPart := make([]string, 0)
	for _, index := range s.indices {
		indexPart = append(indexPart, cleanPathString(index))
	}
	urls += strings.Join(indexPart, ",")

	// Search
	urls += "/_count"

	// Parameters
	params := make(url.Values)
	if s.pretty {
		params.Set("pretty", fmt.Sprintf("%v", s.pretty))
	}
	if len(params) > 0 {
		urls += "?" + params.Encode()
	}

	// Set up a new request
	req, err := s.client.NewRequest("GET", urls)
	if err != nil {
		return 0, err
	}

	if s.debug {
		out, _ := httputil.DumpRequestOut((*http.Request)(req), true)
		fmt.Printf("%s\n", string(out))
	}

	// Get response
	res, err := s.client.c.Do((*http.Request)(req))
	if err != nil {
		return 0, err
	}
	if err := checkResponse(res); err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if s.debug {
		out, _ := httputil.DumpResponse(res, true)
		fmt.Printf("%s\n", string(out))
	}

	ret := new(CountResult)
	if err := json.NewDecoder(res.Body).Decode(ret); err != nil {
		return 0, err
	}

	if ret != nil {
		return ret.Count, nil
	}

	return int64(0), nil
}