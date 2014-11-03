package net

import (
	"fmt"
	"io"
)

type ReadSeeker struct {
	ReadChan     chan int
	ioReadSeeker io.ReadSeeker
}

func NewReadSeeker(ioReadSeeker io.ReadSeeker, readChan chan int) ReadSeeker {
	return ReadSeeker{
		ioReadSeeker: ioReadSeeker,
		ReadChan:     readChan,
	}
}

func (rs ReadSeeker) Read(bytes []byte) (n int, err error) {
	fmt.Println("### ##in read") //never called? droplet not large enough?
	n, err = rs.ioReadSeeker.Read(bytes)
	rs.ReadChan <- n
	return n, err
}

func (rs ReadSeeker) Seek(offset int64, whence int) (int64, error) {
	fmt.Println("### ##in seek")
	return rs.ioReadSeeker.Seek(offset, whence)
}
