package binary

import (
	"fmt"
	"regexp"
	"strings"
)

var repository string

var (
	defaultRepository = "git@github.com:test/some-repo.git"
	uri               = ""
	gitExtention    = regexp.MustCompile(`.git$`)
	azureSshVersion = regexp.MustCompile(`:v\d+`)
)

const (
	ProviderGitHub      = "github.com"
	ProviderAzureDevOps = "dev.azure.com"
	ProviderUnknown     = "unknown Git provider"
)

func Provider(uri string) string {
	if strings.Contains(uri, ProviderGitHub) {
		return ProviderGitHub
	}

	if strings.Contains(uri, ProviderAzureDevOps) {
		return ProviderAzureDevOps
	}

	return ProviderUnknown
}

func init() {
	uri = NormaliseUri(Repository())
}

func Repository() string {
	return getOr(repository, defaultRepository)
}

func Uri() string {
	return uri
}

func NormaliseUri(uri string) string {
	switch Provider(uri) {
	case ProviderAzureDevOps:
		uri = normaliseAzureDevOpsUri(uri)
	default:
		uri = nomaliseGitHubLikeUri(uri)
	}

	return uri
}

func nomaliseGitHubLikeUri(uri string) string {
	isGitProtocol := strings.HasPrefix(uri, "git@")
	if isGitProtocol {
		uri = strings.Replace(uri, ":", "/", 1)
		uri = strings.Replace(uri, "git@", "https://", 1)
	}

	isGitUri := strings.HasPrefix(uri, "git://")
	if isGitUri {
		uri = strings.Replace(uri, "git://", "https://", 1)
	}

	uri = gitExtention.ReplaceAllString(uri, "")

	return uri
}

func normaliseAzureDevOpsUri(uri string) string {
	if strings.Contains(uri, "git@ssh.") {
		lastSlach := strings.LastIndex(uri, "/")
		repository := uri[lastSlach+1:]

		uri = azureSshVersion.ReplaceAllString(uri, "")
		uri = strings.Replace(uri, "git@ssh.", "https://", 1)
		uri = strings.Replace(uri, repository, fmt.Sprintf("_git/%s", repository), 1)
	}

	return uri
}
