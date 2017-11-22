package scp

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"code.cloudfoundry.org/diego-ssh/scp/atime"
)

type TimeMessage struct {
	modificationTime time.Time
	accessTime       time.Time
}

func NewTimeMessage(fileInfo os.FileInfo) *TimeMessage {
	accessTime, err := atime.AccessTime(fileInfo)
	if err != nil {
		accessTime = time.Unix(0, 0)
	}

	return &TimeMessage{
		modificationTime: fileInfo.ModTime(),
		accessTime:       accessTime,
	}
}

func (tm *TimeMessage) ModificationTime() time.Time {
	return tm.modificationTime
}

func (tm *TimeMessage) AccessTime() time.Time {
	return tm.accessTime
}

func (tm *TimeMessage) Send(session *Session) error {
	_, err := fmt.Fprintf(session.stdout, "T%d 0 %d 0\n", tm.modificationTime.Unix(), tm.accessTime.Unix())
	if err != nil {
		return err
	}

	return session.awaitConfirmation()
}

func (tm *TimeMessage) Receive(session *Session) error {
	messageType, err := session.readByte()
	if err != nil {
		return err
	}

	if messageType != byte('T') {
		return fmt.Errorf("unexpected message type: %c", messageType)
	}

	modTimeString, err := session.readString(SPACE)
	if err != nil {
		return err
	}

	modTimeSeconds, err := strconv.ParseUint(modTimeString, 10, 64)
	if err != nil {
		return err
	}

	tm.modificationTime = time.Unix(int64(modTimeSeconds), 0)

	_, err = session.readString(SPACE)
	if err != nil {
		return err
	}

	accessTimeString, err := session.readString(SPACE)
	if err != nil {
		return err
	}

	accessTimeSeconds, err := strconv.ParseUint(accessTimeString, 10, 64)
	if err != nil {
		return err
	}

	tm.accessTime = time.Unix(int64(accessTimeSeconds), 0)

	_, err = session.readString(NEWLINE)
	if err != nil {
		return err
	}

	err = session.sendConfirmation()
	if err != nil {
		return err
	}

	return nil
}
