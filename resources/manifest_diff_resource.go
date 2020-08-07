package resources

type DiffOperation string

const AddOperation DiffOperation = "add"
const ReplaceOperation DiffOperation = "replace"
const RemoveOperation DiffOperation = "remove"

type Diff struct {
	Op    DiffOperation `json:"op"`
	Path  string        `json:"path"`
	Was   interface{}   `json:"was"`
	Value interface{}   `json:"value"`
}

type ManifestDiff struct {
	Diffs []Diff `json:"diff"`
}
