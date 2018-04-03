package cmd

import (
	boshreldir "github.com/cloudfoundry/bosh-cli/releasedir"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type BlobsCmd struct {
	blobsDir boshreldir.BlobsDir
	ui       boshui.UI
}

func NewBlobsCmd(blobsDir boshreldir.BlobsDir, ui boshui.UI) BlobsCmd {
	return BlobsCmd{blobsDir: blobsDir, ui: ui}
}

func (c BlobsCmd) Run() error {
	blobs, err := c.blobsDir.Blobs()
	if err != nil {
		return err
	}

	table := boshtbl.Table{
		Content: "blobs",

		Header: []boshtbl.Header{
			boshtbl.NewHeader("Path"),
			boshtbl.NewHeader("Size"),
			boshtbl.NewHeader("Blobstore ID"),
			boshtbl.NewHeader("Digest"),
		},

		SortBy: []boshtbl.ColumnSort{
			{Column: 0, Asc: true},
		},
	}

	for _, blob := range blobs {
		blobID := blob.BlobstoreID

		if len(blobID) == 0 {
			blobID = "(local)"
		}

		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString(blob.Path),
			boshtbl.NewValueBytes(uint64(blob.Size)),
			boshtbl.NewValueString(blobID),
			boshtbl.NewValueString(blob.SHA1),
		})
	}

	c.ui.PrintTable(table)

	return nil
}
