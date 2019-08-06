package models

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/rtlnl/data-personalization-api/pkg/db"
)

const (
	initVersion = "0.1.0"
	initStage   = STAGED
)

// StageType defines if the model is available for the recommendations or not
type StageType string

const (
	// STAGED instructs the data to be internally available only
	STAGED StageType = "staging"
	// PUBLISHED instructs the data to be available for the recommendations
	PUBLISHED StageType = "production"
)

// Model is the object that acts as container for the metadata of each model
type Model struct {
	PublicationPoint string          `description:"name of the table where the data will live"`
	Campaign         string          `description:"name of the campaign that will be used"`
	Stage            StageType       `description:"defines if the data is available to the clients or not"`
	Version          *semver.Version `description:"internal version of the model. Should follow this pattern Major.Minor"`
	SignalType       string          `description:"definition of the internal signal for composing the key"`
}

// NewModel is invoked when a new model is created in the database.
// Every new model starts with STAGED type and Version 0.1
//
// Definition of the table in Aerospike for models
// setName --> publicationPoint
// key     --> campaign
// bins    --> Version = 0.1
//             Stage = STAGED
//             SignalType = signalType
func NewModel(publicationPoint, campaign, signalType string, ac *db.AerospikeClient) (*Model, error) {

	v, err := semver.NewVersion(initVersion)
	if err != nil {
		return nil, err
	}

	// fill up bins
	bins := make(map[string]interface{})
	bins["version"] = v.String()
	bins["stage"] = initStage
	bins["signal_type"] = signalType

	// create model and fill up metadata
	for k, v := range bins {
		if err := ac.AddOne(publicationPoint, campaign, k, v); err != nil {
			return nil, err
		}
	}

	return &Model{
		PublicationPoint: publicationPoint,
		Campaign:         campaign,
		SignalType:       signalType,
		Stage:            initStage,
		Version:          v,
	}, nil
}

// GetExistingModel returns an already existing model to the caller
func GetExistingModel(publicationPoint, campaign string, ac *db.AerospikeClient) (*Model, error) {
	m, err := ac.GetOne(publicationPoint, campaign)
	if err != nil {
		return nil, err
	}

	// convert version back
	v, err := semver.NewVersion(m.Bins["version"].(string))
	if err != nil {
		return nil, err
	}

	return &Model{
		PublicationPoint: publicationPoint,
		Campaign:         campaign,
		SignalType:       m.Bins["signal_type"].(string),
		Stage:            m.Bins["stage"].(StageType),
		Version:          v,
	}, nil
}

// PublishModel set the model to be available to the clients
// Version: the major value is bumped up
func (m *Model) PublishModel(ac *db.AerospikeClient) error {
	if m.Stage == PUBLISHED {
		return errors.New("model is already published")
	}

	snStaged := m.ComposeSetName()

	// Copy data from staging to published
	recordsStaged, err := ac.GetAllRecords(snStaged)
	if err != nil {
		return err
	}

	// set the published stage
	m.Stage = PUBLISHED
	m.Version.IncMajor()

	snPublished := m.ComposeSetName()
	if err := ac.AddMultipleRecords(snPublished, recordsStaged); err != nil {
		return err
	}

	// delete the staged data of the model
	// Recommendations setName => publicationPoint#campaign#STAGED
	if err := ac.TruncateSet(snStaged); err != nil {
		return err
	}

	return nil
}

// StageModel will set the current model to staging and make it not more available to the clients
// Version: the minor value is bumped up
func (m *Model) StageModel(ac *db.AerospikeClient) error {
	// if the model is already staged no need to perform any operation
	if m.Stage == STAGED {
		return errors.New("model is already STAGED")
	}

	snPublished := m.ComposeSetName()

	// Copy data from published to staging
	recordsPublished, err := ac.GetAllRecords(snPublished)
	if err != nil {
		return err
	}

	// set the staging stage
	m.Stage = STAGED
	m.Version.IncMinor()

	snStaged := m.ComposeSetName()
	if err := ac.AddMultipleRecords(snPublished, recordsPublished); err != nil {
		return err
	}

	// delete the published data of the model
	// Recommendations setName => publicationPoint#campaign#PUBLISHED
	if err := ac.TruncateSet(snStaged); err != nil {
		return err
	}

	return nil
}

// DeleteModel truncate all the data belonging to a model
func (m *Model) DeleteModel(ac *db.AerospikeClient) error {
	// It is not possible to delete models that are published
	if m.Stage == PUBLISHED {
		return errors.New("you cannot delete a model that is PUBLISHED. Change the stage to STAGED first")
	}

	sn := m.ComposeSetName()
	// delete the published data of the model
	// Recommendations setName => publicationPoint#campaign#STAGED
	if err := ac.TruncateSet(sn); err != nil {
		return err
	}

	// reset version back to initial
	v, err := semver.NewVersion(initVersion)
	if err != nil {
		return err
	}
	m.Version = v

	return nil
}

// UpdateSignalType triggers a change in the way the signals are stored
// NOTE: only STAGED models can have a different type of signalType
//       This triggers a deletion of the data because there can be inconsistencies
// Version: the patch value is bumped up
func (m *Model) UpdateSignalType(signalType string, ac *db.AerospikeClient) error {
	if m.Stage == PUBLISHED {
		return errors.New("cannot change signal when model is published")
	}

	// truncate data
	setName := m.ComposeSetName()
	if err := ac.TruncateSet(setName); err != nil {
		return err
	}

	// change signalType
	m.SignalType = signalType
	m.Version.IncPatch()

	return nil
}

// ComposeSetName returns a string with the formatted value of the key we store in Aerospike
func (m *Model) ComposeSetName() string {
	return fmt.Sprintf("%s#%s#%s", m.PublicationPoint, m.Campaign, m.Stage)
}

// ComposeSignalKey returns the actual key composition given a list of signals' values based on the model
func (m *Model) ComposeSignalKey(signals map[string]string) string {
	signalsList := strings.Split(m.SignalType, "_")

	// split and compose key
	var kb strings.Builder
	for _, sKey := range signalsList {
		sVal := signals[sKey]

		kb.WriteString(sVal)
		kb.WriteString("#")
	}

	// remove last occurrence of #
	key := strings.TrimSuffix(kb.String(), "#")
	return key
}
