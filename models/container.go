package models

import (
	"time"

	"github.com/rtlnl/data-personalization-api/pkg/db"
	"github.com/rtlnl/data-personalization-api/utils"
)

const (
	binKeyModels = "models"
)

// Container is used simply as reference to understand where the models are connected to
// Models and Containers are separate entities that have a "fake" relationship
type Container struct {
	PublicationPoint string     `json:"publicationPoint" description:"publication point where the model will be connected to"`
	Campaign         string     `json:"campaign" description:"name for where in a potential place of the internal products, the model will be placed"`
	Models           []string   `json:"models" description:"list of models that are linked to this container"`
	CreatedAt        *time.Time `json:"createdAt" description:"when the container was created"`
}

// NewContainer creates a new container in the database
func NewContainer(publicationPoint, campaign string, models []string, ac *db.AerospikeClient) (*Container, error) {
	// does container exists already then return it to the client
	if c, err := GetExistingContainer(publicationPoint, campaign, ac); c != nil {
		return c, err
	}

	// fill up bins
	bins := make(map[string]interface{})
	bins[binKeyModels] = append([]string{}, models...)

	// create model and fill up metadata
	for k, v := range bins {
		if err := ac.AddOne(publicationPoint, campaign, k, v); err != nil {
			return nil, err
		}
	}

	return &Container{
		PublicationPoint: publicationPoint,
		Campaign:         campaign,
		Models:           bins[binKeyModels].([]string),
	}, nil
}

// GetExistingContainer checks if an existing object already exists or not
func GetExistingContainer(publicationPoint, campaign string, ac *db.AerospikeClient) (*Container, error) {
	c, err := ac.GetOne(publicationPoint, campaign)
	if err != nil {
		return nil, err
	}

	// convert models list back
	ms := utils.ConvertInterfaceToList(c.Bins[binKeyModels])

	return &Container{
		PublicationPoint: publicationPoint,
		Campaign:         campaign,
		Models:           ms,
	}, nil
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
	bins[binKeyModels] = c.Models

	// create model and fill up metadata
	for k, v := range bins {
		if err := ac.AddOne(c.PublicationPoint, c.Campaign, k, v); err != nil {
			return nil, err
		}
	}
	return c, nil
}
