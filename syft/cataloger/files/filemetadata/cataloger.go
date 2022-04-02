package filemetadata

import (
	"github.com/anchore/syft/internal/bus"
	"github.com/anchore/syft/internal/log"
	"github.com/anchore/syft/syft/event"
	"github.com/anchore/syft/syft/file"
	"github.com/wagoodman/go-partybus"
	"github.com/wagoodman/go-progress"
)

type Cataloger struct {
}

func NewCataloger() *Cataloger {
	return &Cataloger{}
}

func (i *Cataloger) Catalog(resolver file.Resolver) (map[file.Coordinates]file.Metadata, error) {
	results := make(map[file.Coordinates]file.Metadata)
	var locations []file.Location
	for location := range resolver.AllLocations() {
		locations = append(locations, location)
	}
	stage, prog := metadataCatalogingProgress(int64(len(locations)))
	for _, location := range locations {
		stage.Current = location.RealPath
		metadata, err := resolver.FileMetadataByLocation(location)
		if err != nil {
			return nil, err
		}

		results[location.Coordinates] = metadata
		prog.N++
	}
	log.Debugf("file metadata cataloger processed %d files", prog.N)
	prog.SetCompleted()
	return results, nil
}

func metadataCatalogingProgress(locations int64) (*progress.Stage, *progress.Manual) {
	stage := &progress.Stage{}
	prog := &progress.Manual{
		Total: locations,
	}

	bus.Publish(partybus.Event{
		Type: event.FileMetadataCatalogerStarted,
		Value: struct {
			progress.Stager
			progress.Progressable
		}{
			Stager:       progress.Stager(stage),
			Progressable: prog,
		},
	})

	return stage, prog
}
