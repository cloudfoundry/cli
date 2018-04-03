package ui

type IndexReporter struct {
	ui UI
}

func NewIndexReporter(ui UI) IndexReporter {
	return IndexReporter{ui: ui}
}

func (r IndexReporter) IndexEntryStartedAdding(type_, desc string) {
	r.ui.BeginLinef("Adding %s '%s'...\n", type_, desc)
}

func (r IndexReporter) IndexEntryFinishedAdding(type_, desc string, err error) {
	if err != nil {
		r.ui.ErrorLinef("Failed adding %s '%s'\n", type_, desc)
	} else {
		r.ui.BeginLinef("Added %s '%s'\n", type_, desc)
	}
}

func (r IndexReporter) IndexEntryDownloadStarted(type_, desc string) {
	r.ui.BeginLinef("-- Started downloading '%s' (%s)\n", type_, desc)
}

func (r IndexReporter) IndexEntryDownloadFinished(type_, desc string, err error) {
	if err != nil {
		r.ui.ErrorLinef("-- Failed downloading '%s' (%s)\n", type_, desc)
	} else {
		r.ui.BeginLinef("-- Finished downloading '%s' (%s)\n", type_, desc)
	}
}

func (r IndexReporter) IndexEntryUploadStarted(type_, desc string) {
	r.ui.BeginLinef("-- Started uploading '%s' (%s)\n", type_, desc)
}

func (r IndexReporter) IndexEntryUploadFinished(type_, desc string, err error) {
	if err != nil {
		r.ui.ErrorLinef("-- Failed uploading '%s' (%s)\n", type_, desc)
	} else {
		r.ui.BeginLinef("-- Finished uploading '%s' (%s)\n", type_, desc)
	}
}
