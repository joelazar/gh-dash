package context

import (
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
)

func TestPRSections_HidesRepoOnlyWithoutRepo(t *testing.T) {
	cfg := config.Config{PRSections: []config.PrsSectionConfig{
		{Title: "mine"},
		{Title: "bots", RepoOnly: true},
	}}
	ctx := &ProgramContext{Config: &cfg}
	require.Len(t, ctx.PRSections(), 1)

	ctx.GHRepo = &repository.Repository{Owner: "onomondo", Name: "eim"}
	require.Len(t, ctx.PRSections(), 2)
}
