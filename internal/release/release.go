// Package release defines URL templates and metadata for terrier's GitHub
// release artifacts, plus the metadata-fetch operation against the GitHub
// API. Centralizing these prevents drift between the daily update check
// and the self-update command.
package release

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

const (
	// BinaryName is the install name. Release assets are named
	// "<BinaryName>-<os>-<arch>".
	BinaryName = "terrier"

	// Alias is the short name installed alongside the binary.
	Alias = "ter"

	// Repo is the GitHub <owner>/<name> slug.
	Repo = "sylophi/" + BinaryName

	// LatestAPI is the GitHub endpoint returning the latest release JSON.
	LatestAPI = "https://api.github.com/repos/" + Repo + "/releases/latest"

	// AcceptHeader is GitHub's recommended Accept value for the API.
	AcceptHeader = "application/vnd.github+json"
)

// statusError explains a failed response.
//
// The one failure worth naming is the rate limit, because the GitHub API
// allows only 60 unauthenticated requests an hour per address and terrier
// has no token to raise that with. A bare "HTTP 403" reads as something
// being broken, when the answer is to run the command again later.
func statusError(resp *http.Response) error {
	if resp.StatusCode == http.StatusForbidden && resp.Header.Get("X-RateLimit-Remaining") == "0" {
		if reset, err := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64); err == nil {
			return fmt.Errorf("GitHub is rate limiting this address. It resets at %s, and `terrier update` will work again then", time.Unix(reset, 0).Format("15:04"))
		}
		return fmt.Errorf("GitHub is rate limiting this address. Try `terrier update` again later")
	}
	return fmt.Errorf("HTTP %d", resp.StatusCode)
}

// AssetURL returns the download URL for a tagged binary release.
func AssetURL(tag, asset string) string {
	return fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", Repo, tag, asset)
}

// AssetName returns the release-asset filename for a given platform suffix
// (e.g., "darwin-arm64").
func AssetName(suffix string) string {
	return BinaryName + "-" + suffix
}

// AssetSuffix returns the platform suffix of the running binary, naming
// the asset the release workflow built for it.
//
// It lives here beside AssetName because the two halves of an asset's
// filename drifting apart is the failure this package exists to prevent,
// and it shows up as a 404 partway through `terrier update`. The same
// matrix is spelled in install.sh and the release workflow, which cannot
// call this.
func AssetSuffix() (string, error) {
	unsupported := fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return "", unsupported
	}
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x64"
	}
	if arch != "arm64" && arch != "x64" {
		return "", unsupported
	}
	return runtime.GOOS + "-" + arch, nil
}

// FetchLatestTag returns the tag_name of the latest GitHub release. The
// context governs the request deadline.
func FetchLatestTag(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, LatestAPI, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", AcceptHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", statusError(resp)
	}
	var data struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	if data.TagName == "" {
		return "", fmt.Errorf("missing tag_name")
	}
	return data.TagName, nil
}
