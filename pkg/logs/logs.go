package logs

import (
	"encoding/json"
	"time"

	"github.com/rtlnl/phoenix/models"
)

// RowLog is the object that will be written in the logs
type RowLog struct {
	PublicationPoint string
	Campaign         string
	SignalID         string
	ItemScores       []models.ItemScore
}

// RecommendationLog is the interface for the different type of logging system
type RecommendationLog interface {
	Write(RowLog) error
}

// CreateLogMessage append extra information to the item score object
func CreateLogMessage(publicationPoint, campaign, signalID string, is models.ItemScore) ([]byte, error) {
	item := make(map[string]string)

	// append timestamp and signalID
	item["timestamp"] = time.Now().Format("2006-01-02 15:04:05")
	item["publicationPoint"] = publicationPoint
	item["campaign"] = campaign
	item["signalId"] = signalID

	// copy the rest
	for k, v := range is {
		item[k] = v
	}

	// create JSON string
	return json.Marshal(item)
}
