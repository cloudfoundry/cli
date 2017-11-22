package scp

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"code.cloudfoundry.org/lager"
)

func (s *secureCopy) SendFile(file *os.File, fileInfo os.FileInfo) error {
	logger := s.session.logger.Session("send-file")
	if fileInfo.IsDir() {
		return errors.New("cannot send a directory")
	}

	if s.session.preserveTimesAndMode {
		timeMessage := NewTimeMessage(fileInfo)
		err := timeMessage.Send(s.session)
		if err != nil {
			return err
		}
	}

	_, err := fmt.Fprintf(s.session.stdout, "C%.4o %d %s\n", fileInfo.Mode()&07777, fileInfo.Size(), fileInfo.Name())
	if err != nil {
		return err
	}

	err = s.session.awaitConfirmation()
	if err != nil {
		return err
	}

	bytesSent, err := io.CopyN(s.session.stdout, file, fileInfo.Size())
	if err != nil {
		return err
	}

	err = s.session.sendConfirmation()
	if err != nil {
		return err
	}

	logger.Info("awaiting-contents-confirmation", lager.Data{"File Size": fileInfo.Size(), "Bytes Sent": bytesSent})
	err = s.session.awaitConfirmation()
	if err != nil {
		logger.Error("failed-contents-confirmation", err)
		return err
	}
	logger.Info("recieved-contents-confirmation")

	return nil
}

func (s *secureCopy) ReceiveFile(path string, pathIsDir bool, timeMessage *TimeMessage) error {
	messageType, err := s.session.readByte()
	if err != nil {
		return err
	}

	if messageType != byte('C') {
		return fmt.Errorf("unexpected message type: %c", messageType)
	}

	fileModeString, err := s.session.readString(SPACE)
	if err != nil {
		return err
	}

	fileMode, err := strconv.ParseUint(fileModeString, 8, 32)
	if err != nil {
		return err
	}

	lengthString, err := s.session.readString(SPACE)
	if err != nil {
		return err
	}

	length, err := strconv.ParseInt(lengthString, 10, 64)
	if err != nil {
		return err
	}

	fileName, err := s.session.readString(NEWLINE)
	if err != nil {
		return err
	}

	err = s.session.sendConfirmation()
	if err != nil {
		return err
	}

	targetPath := path
	if pathIsDir {
		targetPath = filepath.Join(path, fileName)
	}

	targetFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(fileMode))
	if err != nil {
		return err
	}

	_, err = io.CopyN(targetFile, s.session.stdin, length)
	targetFile.Close()
	if err != nil {
		return err
	}

	if s.session.preserveTimesAndMode {
		err := os.Chmod(targetPath, os.FileMode(fileMode))
		if err != nil {
			return err
		}
	}

	// OpenSSH does not check the flag
	if timeMessage != nil {
		err := os.Chtimes(targetPath, timeMessage.accessTime, timeMessage.modificationTime)
		if err != nil {
			return err
		}
	}

	err = s.session.awaitConfirmation()
	if err != nil {
		return err
	}

	err = s.session.sendConfirmation()
	if err != nil {
		return err
	}

	return nil
}
