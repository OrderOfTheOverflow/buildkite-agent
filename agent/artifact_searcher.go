package agent

import (
	"context"
	"time"

	"github.com/buildkite/agent/v3/api"
	"github.com/buildkite/agent/v3/logger"
	"github.com/buildkite/roko"
)

type ArtifactSearcher struct {
	// The logger instance to use
	logger logger.Logger

	// The APIClient that will be used when uploading jobs
	apiClient APIClient

	// The ID of the Build that these artifacts belong to
	buildID string
}

func NewArtifactSearcher(l logger.Logger, ac APIClient, buildID string) *ArtifactSearcher {
	return &ArtifactSearcher{
		logger:    l,
		apiClient: ac,
		buildID:   buildID,
	}
}

func (a *ArtifactSearcher) Search(ctx context.Context, query, scope string, includeRetriedJobs, includeDuplicates bool) ([]*api.Artifact, error) {
	if scope == "" {
		a.logger.Info("Searching for artifacts: \"%s\"", query)
	} else {
		a.logger.Info("Searching for artifacts: \"%s\" within step: \"%s\"", query, scope)
	}

	var artifacts []*api.Artifact

	// Retry on transport errors, a failed search will return 0 artifacts
	err := roko.NewRetrier(
		roko.WithMaxAttempts(10),
		roko.WithStrategy(roko.Constant(5*time.Second)),
	).DoWithContext(ctx, func(*roko.Retrier) error {
		var searchErr error
		artifacts, _, searchErr = a.apiClient.SearchArtifacts(ctx, a.buildID, &api.ArtifactSearchOptions{
			Query:              query,
			Scope:              scope,
			State:              "finished",
			IncludeRetriedJobs: includeRetriedJobs,
			IncludeDuplicates:  includeDuplicates,
		})
		return searchErr
	})

	return artifacts, err
}
