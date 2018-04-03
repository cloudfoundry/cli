package blobstore

type Blobstore interface {
	Get(blobID string) (fileName string, err error)

	CleanUp(fileName string) (err error)

	Create(fileName string) (blobID string, err error)

	Validate() (err error)

	Delete(blobId string) (err error)
}
