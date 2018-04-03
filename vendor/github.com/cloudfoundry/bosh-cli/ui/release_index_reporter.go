package ui

type ReleaseIndexReporter struct {
	ui UI
}

func NewReleaseIndexReporter(ui UI) ReleaseIndexReporter {
	return ReleaseIndexReporter{ui: ui}
}

func (r ReleaseIndexReporter) ReleaseIndexAdded(name, desc string, err error) {
	if err != nil {
		r.ui.ErrorLinef("Failed adding %s release '%s'", name, desc)
	} else {
		r.ui.PrintLinef("Added %s release '%s'", name, desc)
	}
}
