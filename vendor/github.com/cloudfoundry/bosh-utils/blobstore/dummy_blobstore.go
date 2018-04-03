package blobstore

type dummyBlobstore struct{}

func newDummyBlobstore() dummyBlobstore {
	return dummyBlobstore{}
}

func (b dummyBlobstore) Get(blobID string) (string, error) {
	return "", nil
}

func (b dummyBlobstore) CleanUp(fileName string) error {
	return nil
}

func (b dummyBlobstore) Create(fileName string) (string, error) {
	return "", nil
}

func (b dummyBlobstore) Validate() error {
	return nil
}

func (b dummyBlobstore) Delete(blobID string) error {
	return nil
}
