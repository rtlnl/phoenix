package models

// ItemScore is the object containing the recommended item and its score
// Example: {"item":"11","score":"0.6"}
type ItemScore map[string]string

// RecordQueue is the object used to upload the data when coming from S3 with channels
type RecordQueue struct {
	SetName string
	Entry   SingleEntry
}

// SingleEntry is the object used to unmarshal a single JSON line
type SingleEntry struct {
	SignalID    string      `json:"signalID"`
	Recommended []ItemScore `json:"recommended"`
}
