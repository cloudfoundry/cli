package envelope_extensions

import (
	"encoding/binary"
	"fmt"

	"github.com/cloudfoundry/sonde-go/events"
)

const SystemAppId = "system"

type hasAppId interface {
	GetApplicationId() *events.UUID
}

func GetAppId(envelope *events.Envelope) string {
	if envelope.GetEventType() == events.Envelope_LogMessage {
		return envelope.GetLogMessage().GetAppId()
	}

	if envelope.GetEventType() == events.Envelope_ContainerMetric {
		return envelope.GetContainerMetric().GetApplicationId()
	}

	var event hasAppId
	switch envelope.GetEventType() {
	case events.Envelope_HttpStart:
		event = envelope.GetHttpStart()
	case events.Envelope_HttpStop:
		event = envelope.GetHttpStop()
	case events.Envelope_HttpStartStop:
		event = envelope.GetHttpStartStop()
	default:
		return SystemAppId
	}

	uuid := event.GetApplicationId()
	if uuid != nil {
		return formatUUID(uuid)
	}
	return SystemAppId
}

func formatUUID(uuid *events.UUID) string {
	var uuidBytes [16]byte
	binary.LittleEndian.PutUint64(uuidBytes[:8], uuid.GetLow())
	binary.LittleEndian.PutUint64(uuidBytes[8:], uuid.GetHigh())
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuidBytes[0:4], uuidBytes[4:6], uuidBytes[6:8], uuidBytes[8:10], uuidBytes[10:])
}
