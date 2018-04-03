package fakes

type FakeBlobstore struct {
	GetInputs []GetInput
	GetErr    error

	AddInputs []AddInput
	AddBlobID string
	AddErr    error
}

type GetInput struct {
	BlobID          string
	DestinationPath string
}

type AddInput struct {
	SourcePath string
}

func NewFakeBlobstore() *FakeBlobstore {
	return &FakeBlobstore{}
}

func (b *FakeBlobstore) Get(blobID string, destinationPath string) error {
	b.GetInputs = append(b.GetInputs, GetInput{
		BlobID:          blobID,
		DestinationPath: destinationPath,
	})

	return b.GetErr
}

func (b *FakeBlobstore) Add(sourcePath string) (blobID string, err error) {
	b.AddInputs = append(b.AddInputs, AddInput{
		SourcePath: sourcePath,
	})

	return b.AddBlobID, b.AddErr
}
