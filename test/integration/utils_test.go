package integration

import (
	"github.com/anchore/syft/syft/cataloger/packages"
	"github.com/anchore/syft/syft/linux"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/anchore/syft/syft/sbom"

	"github.com/anchore/stereoscope/pkg/imagetest"
	"github.com/anchore/syft/syft/source"
)

func catalogFixtureImage(t *testing.T, fixtureImageName string) (sbom.SBOM, *source.Source) {
	imagetest.GetFixtureImage(t, "docker-archive", fixtureImageName)
	tarPath := imagetest.GetFixtureImageTarPath(t, fixtureImageName)
	userInput := "docker-archive:" + tarPath
	sourceInput, err := source.ParseInput(userInput, "", false)
	require.NoError(t, err)
	theSource, cleanupSource, err := source.New(*sourceInput, nil, nil)
	t.Cleanup(cleanupSource)
	require.NoError(t, err)

	// TODO: this would be better with functional options (after/during API refactor)... this should be replaced
	resolver, err := theSource.FileResolver(source.SquashedScope)
	require.NoError(t, err)
	release := linux.IdentifyRelease(resolver)
	pkgCatalog, relationships, err := packages.Catalog(resolver, release, packages.CatalogersBySourceScheme(theSource.Metadata.Scheme, packages.DefaultSearchConfig())...)
	if err != nil {
		t.Fatalf("failed to catalog image: %+v", err)
	}

	return sbom.SBOM{
		Artifacts: sbom.Artifacts{
			Packages:          pkgCatalog,
			LinuxDistribution: release,
		},
		Relationships: relationships,
		Source:        theSource.Metadata,
		Descriptor: sbom.Descriptor{
			Name:    "syft",
			Version: "v0.42.0-bogus",
			// the application configuration should be persisted here, however, we do not want to import
			// the application configuration in this package (it's reserved only for ingestion by the cmd package)
			Configuration: map[string]string{
				"config-key": "config-value",
			},
		},
	}, theSource
}

func catalogDirectory(t *testing.T, dir string) (sbom.SBOM, *source.Source) {
	userInput := "dir:" + dir
	sourceInput, err := source.ParseInput(userInput, "", false)
	require.NoError(t, err)
	theSource, cleanupSource, err := source.New(*sourceInput, nil, nil)
	t.Cleanup(cleanupSource)
	require.NoError(t, err)

	// TODO: this would be better with functional options (after/during API refactor)
	resolver, err := theSource.FileResolver(source.AllLayersScope)
	require.NoError(t, err)
	release := linux.IdentifyRelease(resolver)
	pkgCatalog, relationships, err := packages.Catalog(resolver, release, packages.CatalogersBySourceScheme(theSource.Metadata.Scheme, packages.DefaultSearchConfig())...)
	if err != nil {
		t.Fatalf("failed to catalog image: %+v", err)
	}

	return sbom.SBOM{
		Artifacts: sbom.Artifacts{
			Packages:          pkgCatalog,
			LinuxDistribution: release,
		},
		Relationships: relationships,
		Source:        theSource.Metadata,
	}, theSource
}
