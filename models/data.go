package models

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
)

var (
	reservedNames = []string{"models", "containers"}
	// used to fast unmarshal json strings
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

// ItemScore is the object containing the recommended item and its score
// Example: {"item":"11","score":"0.6","type":"movie"}
type ItemScore map[string]string

// LineError contains the line number as key and the error message as string
// Example: {"1":"error signal format",2:"error bananas too small"}
type LineError map[string]string

// RecordQueue is the object used to upload the data when coming from S3 with channels
type RecordQueue struct {
	Table string
	Entry SingleEntry
	Error *LineError
}

// SingleEntry is the object used to unmarshal a single JSON line
type SingleEntry struct {
	SignalID    string      `json:"signalId"`
	Recommended []ItemScore `json:"recommended"`
}

// DeserializeItemScoreArray attempts to convert a string into an array of ItemScore
func DeserializeItemScoreArray(s string) ([]ItemScore, error) {
	var isArr []ItemScore
	err := json.Unmarshal([]byte(s), &isArr)
	if err != nil {
		return nil, err
	}
	return isArr, nil
}

// DeserializeLineErrorArray attempts to convert a string into an array of LineError
func DeserializeLineErrorArray(s string) ([]LineError, error) {
	var leArr []LineError
	err := json.Unmarshal([]byte(s), &leArr)
	if err != nil {
		return nil, err
	}
	return leArr, nil
}

// DeserializeSingleEntryArray returns an array of single entries object based on the data preview
func DeserializeSingleEntryArray(preview map[string]string) ([]SingleEntry, error) {
	var seArr []SingleEntry
	for signalID, vals := range preview {
		var is []ItemScore
		err := json.UnmarshalFromString(vals, &is)
		if err != nil {
			log.Error().Msgf("could not deserialize value. error: %s", err.Error())
			continue
		}
		seArr = append(seArr, SingleEntry{
			SignalID:    signalID,
			Recommended: is,
		})
	}
	return seArr, nil
}
