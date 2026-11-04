package repopicker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListLocalRepos(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"smsc", "otaaas"} {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "onomondo", name, ".git"), 0o755))
	}
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "onomondo", "not-a-repo"), 0o755))

	got := ListLocalRepos(map[string]string{
		"onomondo/*":     filepath.Join(dir, "onomondo") + "/*",
		"dlvhdr/gh-dash": "/src/gh-dash",
		":owner/:repo":   "/ignored/:owner/:repo",
		"missing/*":      filepath.Join(dir, "nope") + "/*",
	})

	require.Equal(t, []Repo{
		{Name: "dlvhdr/gh-dash", Path: "/src/gh-dash"},
		{Name: "onomondo/otaaas", Path: filepath.Join(dir, "onomondo", "otaaas")},
		{Name: "onomondo/smsc", Path: filepath.Join(dir, "onomondo", "smsc")},
	}, got)
}
