package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofurry/gofurry-nav-backend/common"
)

const (
	duckDuckGoEndpoint        = "https://duckduckgo.com/ac/"
	searchSuggestTimeout      = 3 * time.Second
	searchSuggestMaxBodyBytes = 64 * 1024
)

type duckDuckGoProvider struct {
	client    *http.Client
	configErr error
}

// Construct once per Search service. An invalid configured proxy fails closed.
func newDuckDuckGoProvider(rawProxy string) *duckDuckGoProvider {
	provider := &duckDuckGoProvider{}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil // Empty explicit configuration means direct, not environment proxy discovery.
	if rawProxy != "" {
		proxy, err := url.Parse(strings.TrimSpace(rawProxy))
		if err != nil || proxy.Hostname() == "" || proxy.Fragment != "" || proxy.RawQuery != "" || (proxy.Path != "" && proxy.Path != "/") ||
			(proxy.Scheme != "http" && proxy.Scheme != "https" && proxy.Scheme != "socks5" && proxy.Scheme != "socks5h") {
			provider.configErr = errors.New("suggestion proxy configuration invalid")
			return provider
		}
		transport.Proxy = http.ProxyURL(proxy)
	}
	provider.client = &http.Client{
		Transport:     transport,
		Timeout:       searchSuggestTimeout,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	return provider
}

func (provider *duckDuckGoProvider) Fetch(ctx context.Context, query string) ([]string, error) {
	if provider.configErr != nil {
		return nil, provider.configErr
	}
	query = normalizeSuggestionQuery(query)
	if query == "" {
		return []string{}, nil
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, duckDuckGoEndpoint+"?"+url.Values{"q": {query}}.Encode(), nil)
	if err != nil {
		return nil, errors.New("suggestion request invalid")
	}
	request.Header.Set("User-Agent", common.USER_AGENT)
	request.Header.Set("Accept-Language", common.ACCEPT_LANGUAGE)
	request.Header.Set("Accept", "application/json")
	response, err := provider.client.Do(request)
	if err != nil {
		// Do not expose proxy addresses, credentials or query-bearing transport errors.
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, errors.New("suggestion provider transport unavailable")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, errors.New("suggestion provider returned non-success status")
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, searchSuggestMaxBodyBytes+1))
	if err != nil {
		return nil, errors.New("suggestion provider body unavailable")
	}
	if len(body) > searchSuggestMaxBodyBytes {
		return nil, errors.New("suggestion provider body exceeds limit")
	}
	var items []struct {
		Phrase *string `json:"phrase"`
	}
	if err := json.Unmarshal(body, &items); err != nil || items == nil {
		return nil, errors.New("suggestion provider response invalid")
	}
	phrases := make([]string, 0, len(items))
	for _, item := range items {
		if item.Phrase == nil {
			return nil, errors.New("suggestion provider phrase missing")
		}
		phrases = append(phrases, *item.Phrase)
	}
	return sanitizeSuggestions(phrases), nil
}
