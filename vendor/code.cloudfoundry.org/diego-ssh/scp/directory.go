package scp

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

func (s *secureCopy) SendDirectory(dir string, dirInfo os.FileInfo) error {
	return s.sendDirectory(dir, dirInfo)
}

func (s *secureCopy) ReceiveDirectory(dir string, timeMessage *TimeMessage) error {
	messageType, err := s.session.readByte()
	if err != nil {
		return err
	}

	if messageType != byte('D') {
		return fmt.Errorf("unexpected message type: %c", messageType)
	}

	dirModeString, err := s.session.readString(SPACE)
	if err != nil {
		return err
	}

	dirMode, err := strconv.ParseUint(dirModeString, 8, 32)
	if err != nil {
		return err
	}

	// Length field is ignored
	_, err = s.session.readString(SPACE)
	if err != nil {
		return err
	}

	dirName, err := s.session.readString(NEWLINE)
	if err != nil {
		return err
	}

	err = s.session.sendConfirmation()
	if err != nil {
		return err
	}

	targetPath := filepath.Join(dir, dirName)
	_, err = os.Stat(dir)
	if os.IsNotExist(err) {
		targetPath = dir
	}

	targetInfo, err := os.Stat(targetPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		err = os.Mkdir(targetPath, os.FileMode(dirMode))
		if err != nil {
			return err
		}
	} else if !targetInfo.Mode().IsDir() {
		return fmt.Errorf("target exists and is not a directory: %q", dirName)
	}

	err = s.processDirectoryMessages(targetPath)
	if err != nil {
		return err
	}

	message, err := s.session.readString(NEWLINE)
	if message != "E" {
		return fmt.Errorf("unexpected message type: %c", messageType)
	}

	if timeMessage != nil && s.session.preserveTimesAndMode {
		err := os.Chtimes(targetPath, timeMessage.accessTime, timeMessage.modificationTime)
		if err != nil {
			return err
		}
	}

	err = s.session.sendConfirmation()
	if err != nil {
		return err
	}

	return nil
}

func (s *secureCopy) processDirectoryMessages(dirPath string) error {
	for {
		messageType, err := s.session.peekByte()
		if err != nil {
			return err
		}

		var timeMessage *TimeMessage
		if messageType == 'T' && s.session.preserveTimesAndMode {
			timeMessage = &TimeMessage{}
			err := timeMessage.Receive(s.session)
			if err != nil {
				return err
			}

			messageType, err = s.session.peekByte()
			if err != nil {
				return err
			}
		}

		switch messageType {
		case 'D':
			err := s.ReceiveDirectory(dirPath, timeMessage)
			if err != nil {
				return err
			}
		case 'C':
			err := s.ReceiveFile(dirPath, true, timeMessage)
			if err != nil {
				return err
			}
		case 'E':
			if timeMessage != nil {
				return fmt.Errorf("unexpected message type: %c", messageType)
			}
			return nil
		default:
			return fmt.Errorf("unexpected message type: %c", messageType)
		}
	}
}

func (s *secureCopy) sendDirectory(dirname string, directoryInfo os.FileInfo) error {
	if s.session.preserveTimesAndMode {
		timeMessage := NewTimeMessage(directoryInfo)
		err := timeMessage.Send(s.session)
		if err != nil {
			return err
		}
	}

	_, err := fmt.Fprintf(s.session.stdout, "D%.4o 0 %s\n", directoryInfo.Mode()&07777, directoryInfo.Name())
	if err != nil {
		return err
	}

	err = s.session.awaitConfirmation()
	if err != nil {
		return err
	}

	fileInfos, err := ioutil.ReadDir(dirname)
	if err != nil {
		return err
	}

	for _, fileInfo := range fileInfos {
		source := filepath.Join(dirname, fileInfo.Name())
		err := s.send(source)
		if err != nil {
			// Ignore error, probably log
		}
	}

	_, err = fmt.Fprintf(s.session.stdout, "E\n")
	if err != nil {
		return err
	}

	err = s.session.awaitConfirmation()
	if err != nil {
		return err
	}

	return nil
}
