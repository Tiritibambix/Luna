package protocols

import (
	"encoding/json"
	"fmt"
	"luna-backend/constants"
	"luna-backend/protocols/caldav"
	"luna-backend/protocols/google"
	"luna-backend/protocols/ical"
	"luna-backend/types"
)

func EmptySourceSettingsByType(sourceType string) (types.SourceSettings, error) {
	switch sourceType {
	case constants.SourceCaldav:
		return &caldav.CaldavSourceSettings{}, nil
	case constants.SourceIcal:
		return &ical.IcalSourceSettings{}, nil
	case constants.SourceGoogle:
		return &google.GoogleSourceSettings{}, nil
	default:
		return nil, fmt.Errorf("unknown source type: %v", sourceType)
	}
}

func SourceSettingsFromJson(sourceType string, rawBody json.RawMessage) (types.SourceSettings, error) {
	sourceSettings, err := EmptySourceSettingsByType(sourceType)
	if err != nil {
		return nil, err
	}

	// Try to unmarshal
	err = json.Unmarshal(rawBody, sourceSettings)
	if err != nil {
		return nil, err
	}

	return sourceSettings, nil
}
