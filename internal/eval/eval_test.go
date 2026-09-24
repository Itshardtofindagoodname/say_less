package eval

import (
	"strings"
	"testing"
)

func TestWrapForLiveReload(t *testing.T) {
	interp := New()
	interp.EnableLiveReload(true)

	plain := interp.wrapForLiveReload(`{"message":"It works"}`)
	if !strings.Contains(plain, "/__sale_reload") {
		t.Error("Expected live-reload script in wrapped response")
	}
	if !strings.Contains(plain, "<pre>") {
		t.Error("Expected plain-text response wrapped in <pre>")
	}
	if !strings.Contains(plain, "It works") {
		t.Error("Expected original body preserved in wrapped response")
	}

	htmlBody := "<html><head></head><body><p>hi</p></body></html>"
	withScript := interp.wrapForLiveReload(htmlBody)
	if !strings.HasSuffix(withScript[:strings.Index(withScript, "</body>")+7], "</body>") {
		t.Error("Expected HTML response to keep closing body tag")
	}
	if !strings.Contains(withScript, "/__sale_reload") {
		t.Error("Expected live-reload script injected into HTML response")
	}
	if strings.Contains(withScript, "<pre>") {
		t.Error("Expected HTML response not to be wrapped in <pre>")
	}
}
