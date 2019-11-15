package models

import (
	"fmt"

	"github.com/rtlnl/phoenix/pkg/db"
	"github.com/rtlnl/phoenix/utils"
)

const (
	setNameContainers = "containers"
)

// Container is used simply as reference to understand where the models are connected to
// Models and Containers are separate entities that have a "fake" relationship
type Container struct {
	PublicationPoint string   `json:"publicationPoint" description:"publication point where the model will be connected to"`
	Campaign         string   `json:"campaign" description:"name for where in a potential place of the internal products, the model will be placed"`
	Models           []string `json:"models" description:"list of models that are linked to this container"`
}

// NewContainer creates a new container in the database
// SetName --> containers
// Key     --> publicationPoint
// Bins    --> campaing => [model_1, model_2, ..., model_n]
func NewContainer(publicationPoint, campaign string, models []string, ac *db.AerospikeClient) (*Container, error) {
	// does container exists already then return it to the client
	if c, err := GetExistingContainer(publicationPoint, campaign, ac); c != nil || err != nil {
		return c, err
	}

	// check if models exist
	if len(models) > 0 {
		for _, m := range models {
			if m != "" {
				if m, _ := GetExistingModel(m, ac); m == nil {
					return nil, fmt.Errorf("model with name %s not found", m)
				}
			}
		}
	}

	// otherwise fill up bins with the new campaign and models
	bins := make(map[string]interface{})
	bins[campaign] = append([]string{}, models...)

	// create model and fill up metadata
	for k, v := range bins {
		if err := ac.AddOne(setNameContainers, publicationPoint, k, v); err != nil {
			return nil, err
		}
	}

	return &Container{
		PublicationPoint: publicationPoint,
		Campaign:         campaign,
		Models:           bins[campaign].([]string),
	}, nil
}

// GetExistingContainer checks if an existing object already exists or not
func GetExistingContainer(publicationPoint, campaign string, ac *db.AerospikeClient) (*Container, error) {
	if c, _ := ac.GetOne(setNameContainers, publicationPoint); c != nil {
		models := utils.ConvertInterfaceToList(c.Bins[campaign])
		return &Container{
			PublicationPoint: publicationPoint,
			Campaign:         campaign,
			Models:           models,
		}, nil
	}
	return nil, nil
}

// DeleteContainer deletes the content of the container by truncating the PublicationPoint (aka setName)
func (c *Container) DeleteContainer(ac *db.AerospikeClient) error {
	// truncate the publication point and its data
	return ac.TruncateSet(c.PublicationPoint)
}

// LinkModel append the models inside Aerospike structure
func (c *Container) LinkModel(models []string, ac *db.AerospikeClient) (*Container, error) {
	// append models in the current container
	tmp := append(c.Models, models...)
	c.Models = utils.RemoveEmptyValueInSlice(tmp)

	// fill up bins
	bins := make(map[string]interface{})
	bins[c.Campaign] = c.Models

	// create model and fill up metadata
	for k, v := range bins {
		if err := ac.AddOne(setNameContainers, c.PublicationPoint, k, v); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// UnlinkModel removes the model from the list of models linked to the container
func (c *Container) UnlinkModel(model string, ac *db.AerospikeClient) (*Container, error) {
	if !utils.StringInSlice(model, c.Models) {
		return c, nil
	}

	// append models in the current container
	tmp := utils.RemoveElemFromSlice(model, c.Models)
	c.Models = utils.RemoveEmptyValueInSlice(tmp)

	// fill up bins
	bins := make(map[string]interface{})
	bins[c.Campaign] = c.Models

	// create model and fill up metadata
	for k, v := range bins {
		if err := ac.AddOne(setNameContainers, c.PublicationPoint, k, v); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// GetAllContainers returns all the containers in the database
func GetAllContainers(ac *db.AerospikeClient) ([]*Container, error) {
	var containers []*Container
	records, err := ac.GetAllRecords(setNameContainers)
	if err != nil {
		return nil, err
	}

	for record := range records.Results() {
		key := record.Record.Key.Value().String()
		bins := record.Record.Bins
		for campaign, models := range bins {
			containers = append(containers, &Container{
				PublicationPoint: key,
				Campaign:         campaign,
				Models:           utils.ConvertInterfaceToList(models),
			})
		}
	}
	return containers, nil
}
