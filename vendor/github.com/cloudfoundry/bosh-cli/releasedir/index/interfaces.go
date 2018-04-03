package index

//go:generate counterfeiter . Index

type Index interface {
	Find(name, version string) (string, string, error)
	Add(name, version, path, sha1 string) (string, string, error)
}

//go:generate counterfeiter . IndexBlobs

type IndexBlobs interface {
	Get(name, blobID, sha1 string) (string, error)
	Add(name, path, sha1 string) (string, string, error)
}

//go:generate counterfeiter . Reporter

type Reporter interface {
	IndexEntryStartedAdding(type_, desc string)
	IndexEntryFinishedAdding(type_, desc string, err error)

	IndexEntryDownloadStarted(type_, desc string)
	IndexEntryDownloadFinished(type_, desc string, err error)

	IndexEntryUploadStarted(type_, desc string)
	IndexEntryUploadFinished(type_, desc string, err error)
}
