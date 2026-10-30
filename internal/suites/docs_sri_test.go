package suites

import (
	"crypto/sha512"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/michaelolof/gofi"
)

var pinnedAsset = regexp.MustCompile(`(?:src|href)="(https://[^"]+)"[^>]*?integrity="sha384-([^"]+)"`)

// TestDocsAssetIntegrity fetches every asset the docs templates pin and checks
// it against its integrity hash; a wrong hash makes the browser block the
// asset and the docs page renders blank. It needs network access, so it runs
// only with GOFI_DOCS_SRI_IT=1. Run it whenever a pinned version changes, and
// load each UI in a browser as well: a correct hash does not prove the page
// renders.
func TestDocsAssetIntegrity(t *testing.T) {
	if os.Getenv("GOFI_DOCS_SRI_IT") != "1" {
		t.Skip("set GOFI_DOCS_SRI_IT=1 to verify pinned docs assets against their CDNs")
	}
	templates := map[string]gofi.DocsUiTemplate{
		"swagger":   gofi.SwaggerTemplate(),
		"scalar":    gofi.ScalarTemplate(nil),
		"redoc":     gofi.RedoclyTemplate(),
		"rapidoc":   gofi.RapidDoc(),
		"stoplight": gofi.StopLight(),
	}
	client := &http.Client{Timeout: time.Minute}
	for name, tpl := range templates {
		assets := pinnedAsset.FindAllStringSubmatch(string(tpl.HTML("/q/openapi")), -1)
		if len(assets) == 0 {
			t.Errorf("%s: no pinned assets found", name)
		}
		for _, a := range assets {
			url, want := a[1], a[2]
			resp, err := client.Get(url)
			if err != nil {
				t.Errorf("%s: %s: %v", name, url, err)
				continue
			}
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil || resp.StatusCode != http.StatusOK {
				t.Errorf("%s: %s: status %d, %v", name, url, resp.StatusCode, err)
				continue
			}
			sum := sha512.Sum384(body)
			if got := base64.StdEncoding.EncodeToString(sum[:]); got != want {
				t.Errorf("%s: %s: integrity sha384-%s, served file is sha384-%s", name, url, want, got)
			}
		}
	}
}
