package models

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rtlnl/phoenix/pkg/db"
	"github.com/rtlnl/phoenix/utils"
)

const (
	tableModels = "models"
)

// Model is the object that acts as container for the metadata of each model
type Model struct {
	Name         string   `json:"name" description:"name of the model that will be used"`
	SignalOrder  []string `json:"signalOrder" description:"list of ordered signals"`
	Concatenator string   `json:"concatenator" description:"character used as concatenator for SignalOrder {'|','#','_','-'}"`
}

// NewModel is invoked when a new model is created in the database.
func NewModel(name, concatenator string, signalOrder []string, dbc db.DB) (Model, error) {
	// if the model exists return error to the client
	if ModelExists(name, dbc) {
		return Model{}, fmt.Errorf("model with name %s already exists", name)
	}

	// validate model's parameters
	if len(signalOrder) > 1 && concatenator == "" {
		return Model{}, errors.New("multiple signals are being specified but no concatenator. concatenator is missing")
	}

	// check if name used is reserved
	if utils.StringInSlice(name, reservedNames) {
		return Model{}, fmt.Errorf("cannot use %s as name. this name is reserved", name)
	}

	// create model object
	model := Model{
		Name:         name,
		SignalOrder:  signalOrder,
		Concatenator: concatenator,
	}
	// serialize it
	serialized, err := utils.SerializeObject(model)
	if err != nil {
		return Model{}, fmt.Errorf("could not serialize model. error: %s", err.Error())
	}
	// add to the models table
	if err := dbc.AddOne(tableModels, name, serialized); err != nil {
		return Model{}, err
	}
	return model, nil
}

// ModelExists checks if the model exists in the database
func ModelExists(name string, dbc db.DB) bool {
	if _, err := dbc.GetOne(tableModels, name); err != nil {
		return false
	}
	return true
}

// GetModel returns an already existing model to the caller
func GetModel(name string, dbc db.DB) (Model, error) {
	if !ModelExists(name, dbc) {
		return Model{}, fmt.Errorf("model with name %s not found", name)
	}
	// if it fails then return the error directly
	m, err := dbc.GetOne(tableModels, name)
	if err != nil {
		return Model{}, err
	}
	return DeserializeModel(m)
}

// DeleteModel truncate all the data belonging to a model
func (m *Model) DeleteModel(dbc db.DB) error {
	// remove from models
	if err := dbc.DeleteOne(tableModels, m.Name); err != nil {
		return fmt.Errorf("error in removing the model. error: %s", err.Error())
	}
	// remove the whole dataset
	if err := dbc.DropTable(m.Name); err != nil {
		return fmt.Errorf("error in deleting the data of the model. error: %s", err.Error())
	}
	// remove from containers
	containers, err := GetAllContainers(dbc)
	if err != nil {
		return fmt.Errorf("error in retrieving the containers. error: %s", err.Error())
	}
	for _, container := range containers {
		// update models
		tmp := container.Models
		container.Models = utils.RemoveElemFromSlice(m.Name, tmp)
		// store it back
		ser, err := utils.SerializeObject(container)
		if err != nil {
			return fmt.Errorf("failed to serialize container. error: %s", err.Error())
		}
		if err := dbc.AddOne(tableContainers, ContainerUniqueName(container.PublicationPoint, container.Campaign), ser); err != nil {
			return fmt.Errorf("failed to insert container in database. error: %s", err.Error())
		}
	}
	return nil
}

// UpdateSignalOrder triggers a change in the way the signals are stored
func (m *Model) UpdateSignalOrder(signalOrder []string, dbc db.DB) error {
	// delete the data
	if err := dbc.DropTable(m.Name); err != nil {
		return fmt.Errorf("error in deleting the data of the model. error: %s", err.Error())
	}
	// change signalType
	m.SignalOrder = signalOrder
	// serialize model
	model, err := utils.SerializeObject(m)
	if err != nil {
		return fmt.Errorf("failed serialization of the model. error: %s", err.Error())
	}
	// store model
	if err := dbc.AddOne(tableModels, m.Name, model); err != nil {
		return fmt.Errorf("error in storing the signalOrder in database. error: %s", err.Error())
	}
	return nil
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
		if len(r) == 0 {
			return false
		}
		return c == r[0]
	})
	return len(m.SignalOrder) == len(res)
}

// GetAllModels is a convenient functions to get all the models from DB
func GetAllModels(dbc db.DB) ([]Model, error) {
	var models []Model
	records, err := dbc.GetAllRecords(tableModels)
	if err != nil {
		return nil, fmt.Errorf("error in returning all the models from the database. error: %s", err.Error())
	}
	// iterate through all the records
	for modelName := range records {
		m, err := GetModel(modelName, dbc)
		if err != nil {
			return nil, err
		}
		models = append(models, m)
	}
	return models, nil
}

// GetDataPreview returns a limited amount of data as preview for a single model
func (m *Model) GetDataPreview(dbc db.DB) (map[string]string, error) {
	records, err := dbc.GetAllRecords(m.Name)
	if err != nil {
		return nil, fmt.Errorf("error in returning the data preview from the database. error: %s", err.Error())
	}
	return records, nil
}

// DeserializeModel takes a JSON string in input and try to convert it to a Model object
func DeserializeModel(m string) (Model, error) {
	var model Model
	err := json.Unmarshal([]byte(m), &model)
	if err != nil {
		return Model{}, err
	}
	return model, nil
}
