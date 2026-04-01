package app

import (
	"encoding/json"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestBuildTxFilterQueryRejectsInjectedSender(t *testing.T) {
	req := httptest.NewRequest("GET", "/txs?sender=x%27%20AND%20tx.height%3E0", nil)
	if _, err := buildTxFilterQuery(req); err == nil {
		t.Fatal("expected injected sender to be rejected")
	}
}

func TestBuildTxFilterQueryAllowsValidInputs(t *testing.T) {
	req := httptest.NewRequest("GET", "/txs?sender=0x1111111111111111111111111111111111111111&type=register_agent", nil)
	query, err := buildTxFilterQuery(req)
	if err != nil {
		t.Fatalf("expected valid query, got %v", err)
	}
	expected := "message.sender='0x1111111111111111111111111111111111111111' AND message.action='register_agent'"
	if query != expected {
		t.Fatalf("unexpected query: %s", query)
	}
}

func TestWriteCacheRespectsCapacityLimit(t *testing.T) {
	p := &publicAPI{cache: make(map[string]publicAPICacheEntry)}
	payload := json.RawMessage(`{"ok":true}`)

	for i := 0; i < publicAPIMaxCacheEntries+100; i++ {
		p.writeCache(strconv.Itoa(i), time.Minute, payload, time.Now().UTC())
	}

	if got := len(p.cache); got > publicAPIMaxCacheEntries {
		t.Fatalf("cache size exceeded limit: %d", got)
	}
}
