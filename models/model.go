package models

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/rtlnl/data-personalization-api/pkg/db"
	"github.com/rtlnl/data-personalization-api/utils"
)

const (
	setNameComposition = "%s#%s"
	initVersion        = "0.1.0"
	initStage          = STAGED
	binKeyStage        = "stage"
	binKeyVersion      = "version"
	binKeySignalOrder  = "signal_order"
	binKeyConcatenator = ""
)

// StageType defines if the model is available for the recommendations or not
type StageType string

const (
	// STAGED instructs the data to be internally available only
	STAGED StageType = "STAGED"
	// PUBLISHED instructs the data to be available for the recommendations
	PUBLISHED StageType = "PUBLISHED"
)

// Model is the object that acts as container for the metadata of each model
type Model struct {
	PublicationPoint string          `json:"publicationPoint" description:"name of the table where the data will live"`
	Campaign         string          `json:"campaign" description:"name of the campaign that will be used"`
	Stage            StageType       `json:"stage" description:"defines if the data is available to the clients or not"`
	Version          *semver.Version `json:"version" description:"internal version of the model. Should follow this pattern Major.Minor"`
	SignalOrder      []string        `json:"signalOrder" description:"list of ordered signals"`
	Concatenator     string          `json:"concatenator" description:"character used as concatenator for SignalOrder {'|','#','_','-'}"`
}

// NewModel is invoked when a new model is created in the database.
// Every new model starts with STAGED type and Version 0.1
//
// Definition of the table in Aerospike for models
// setName --> publicationPoint
// key     --> campaign
// bins    --> Version = 0.1
//             Stage = STAGED
//             SignalOrder = signalOrder
func NewModel(publicationPoint, campaign, concatenator string, signalOrder []string, ac *db.AerospikeClient) (*Model, error) {
	v, err := semver.NewVersion(initVersion)
	if err != nil {
		return nil, err
	}

	// does model exists already then return it to the client
	if m, err := GetExistingModel(publicationPoint, campaign, ac); m != nil {
		return m, err
	}

	// fill up bins
	bins := make(map[string]interface{})
	bins["version"] = v.String()
	bins["stage"] = initStage
	bins["signal_order"] = signalOrder
	bins["concatenator"] = concatenator

	// create model and fill up metadata
	for k, v := range bins {
		if err := ac.AddOne(publicationPoint, campaign, k, v); err != nil {
			return nil, err
		}
	}

	return &Model{
		PublicationPoint: publicationPoint,
		Campaign:         campaign,
		SignalOrder:      signalOrder,
		Stage:            initStage,
		Version:          v,
		Concatenator:     concatenator,
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

	// read string and convert to enum
	stg := m.Bins["stage"].(string)

	return &Model{
		PublicationPoint: publicationPoint,
		Campaign:         campaign,
		SignalOrder:      utils.ConvertInterfaceToList(m.Bins["signal_order"]),
		Stage:            StageType(stg),
		Version:          v,
		Concatenator:     m.Bins["concatenator"].(string),
	}, nil
}

// PublishModel set the model to be available to the clients
// Version: the major value is bumped up
func (m *Model) PublishModel(ac *db.AerospikeClient) error {
	if m.IsPublished() {
		return errors.New("model is already published")
	}

	// set the published stage
	m.Stage = PUBLISHED
	m.Version.IncMajor()

	// store model stage
	if err := ac.AddOne(m.PublicationPoint, m.Campaign, binKeyStage, m.Stage); err != nil {
		return err
	}

	// store new version
	if err := ac.AddOne(m.PublicationPoint, m.Campaign, binKeyVersion, m.Version.String()); err != nil {
		return err
	}

	return nil
}

// StageModel will set the current model to staging and make it not more available to the clients
// Version: the minor value is bumped up
func (m *Model) StageModel(ac *db.AerospikeClient) error {
	// if the model is already staged no need to perform any operation
	if m.IsStaged() {
		return errors.New("model is already STAGED")
	}

	// set the staging stage
	m.Stage = STAGED
	m.Version.IncMinor()

	// store model stage
	if err := ac.AddOne(m.PublicationPoint, m.Campaign, binKeyStage, m.Stage); err != nil {
		return err
	}

	// store new version
	if err := ac.AddOne(m.PublicationPoint, m.Campaign, binKeyVersion, m.Version.String()); err != nil {
		return err
	}

	return nil
}

// DeleteModel truncate all the data belonging to a model
func (m *Model) DeleteModel(ac *db.AerospikeClient) error {
	// It is not possible to delete models that are published
	if m.IsPublished() {
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

	// store new version
	if err := ac.AddOne(m.PublicationPoint, m.Campaign, binKeyVersion, m.Version.String()); err != nil {
		return err
	}

	return nil
}

// UpdateSignalOrder triggers a change in the way the signals are stored
// NOTE: only STAGED models can have a different type of signalOrder
//       This triggers a deletion of the data because there can be inconsistencies
// Version: the patch value is bumped up
func (m *Model) UpdateSignalOrder(signalOrder []string, ac *db.AerospikeClient) error {
	if m.IsPublished() {
		return errors.New("cannot change signal when model is published")
	}

	// truncate data
	setName := m.ComposeSetName()
	if err := ac.TruncateSet(setName); err != nil {
		return err
	}

	// change signalType
	m.SignalOrder = signalOrder
	m.Version.IncPatch()

	// store model signalOrder
	if err := ac.AddOne(m.PublicationPoint, m.Campaign, binKeySignalOrder, m.SignalOrder); err != nil {
		return err
	}

	// store new version
	if err := ac.AddOne(m.PublicationPoint, m.Campaign, binKeyVersion, m.Version.String()); err != nil {
		return err
	}

	return nil
}

// ComposeSetName returns a string with the formatted value of the key we store in Aerospike
func (m *Model) ComposeSetName() string {
	return fmt.Sprintf(setNameComposition, m.PublicationPoint, m.Campaign)
}

// IsStaged determines if the model is in STAGED mode
func (m *Model) IsStaged() bool {
	return m.Stage == STAGED
}

// IsPublished determines if the model is in PUBLISHED mode
func (m *Model) IsPublished() bool {
	return m.Stage == PUBLISHED
}

// RequireSignalFormat checks if it is required to check the signal format
func (m *Model) RequireSignalFormat() bool {
	if len(m.SignalOrder) > 1 && m.Concatenator != "" {
		return true
	}
	return false
}

// CorrectSignalFormat checks that the signal format is correct
func (m *Model) CorrectSignalFormat(s string) bool {
	res := strings.FieldsFunc(s, func(c rune) bool {
		r := []rune(m.Concatenator)
		return c == r[0]
	})
	return len(m.SignalOrder) == len(res)
}
