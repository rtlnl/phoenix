package models

import (
	"encoding/json"
	"fmt"
	"github.com/rtlnl/phoenix/pkg/db"
	"github.com/rtlnl/phoenix/utils"
)

const (
	tableContainers           = "containers"
	uniqueContainerNameFormat = "%s:%s"
)

// Container is used simply as reference to understand where the models are connected to
// Models and Containers are separate entities that have a "fake" relationship
type Container struct {
	PublicationPoint string   `json:"publicationPoint" description:"publication point where the model will be connected to"`
	Campaign         string   `json:"campaign" description:"name for where in a potential place of the internal products, the model will be placed"`
	Models           []string `json:"models,omitempty" description:"list of models that are linked to this container"`
}

// NewContainer creates a new container in the database
func NewContainer(publicationPoint, campaign string, models []string, dbc db.DB) (Container, error) {
	// if it exists already return error to the client
	if ContainerExists(publicationPoint, campaign, dbc) {
		return Container{}, fmt.Errorf("container with publication point %s and campaign %s already exists", publicationPoint, campaign)
	}
	// check if models exist
	if len(models) > 0 {
		for _, m := range models {
			if m != "" {
				if !ModelExists(m, dbc) {
					return Container{}, fmt.Errorf("model with name %s not found", m)
				}
			}
		}
	}
	// create container object
	container := Container{
		PublicationPoint: publicationPoint,
		Campaign:         campaign,
		Models:           models,
	}
	// serialize container
	serialized, err := utils.SerializeObject(container)
	if err != nil {
		return Container{}, fmt.Errorf("could not serialize container. error: %s", err.Error())
	}
	// store in db
	err = dbc.AddOne(tableContainers, ContainerUniqueName(publicationPoint, campaign), serialized)
	if err != nil {
		return Container{}, err
	}
	return container, nil
}

// GetContainer checks if an existing object already exists or not
func GetContainer(publicationPoint, campaign string, dbc db.DB) (Container, error) {
	if !ContainerExists(publicationPoint, campaign, dbc) {
		return Container{}, fmt.Errorf("container with publication point %s and campaign %s not found", publicationPoint, campaign)
	}
	// retrieve from db
	c, err := dbc.GetOne(tableContainers, ContainerUniqueName(publicationPoint, campaign))
	if err != nil {
		return Container{}, err
	}
	// deserialize container
	return DeserializeContainer(c)
}

// ContainerExists checks if the container is actually in the database
func ContainerExists(publicationPoint, campaign string, dbc db.DB) bool {
	if _, err := dbc.GetOne(tableContainers, ContainerUniqueName(publicationPoint, campaign)); err != nil {
		return false
	}
	return true
}

// DeleteContainer deletes the content of the container by truncating the PublicationPoint (aka setName)
func (c *Container) DeleteContainer(dbc db.DB) error {
	// delete from the containers table
	return dbc.DeleteOne(tableContainers, ContainerUniqueName(c.PublicationPoint, c.Campaign))
}

// LinkModel append the models inside DB structure
func (c *Container) LinkModel(models []string, dbc db.DB) error {
	// append models in the current container
	tmp := append(c.Models, models...)
	// update models property
	c.Models = utils.RemoveEmptyValueInSlice(tmp)
	// serialize object
	container, err := utils.SerializeObject(c)
	if err != nil {
		return fmt.Errorf("failed to serialize contianer. error: %s", err.Error())
	}
	// update database
	err = dbc.AddOne(tableContainers, ContainerUniqueName(c.PublicationPoint, c.Campaign), container)
	if err != nil {
		return fmt.Errorf("failed to insert container into db. error: %s", err.Error())
	}
	return nil
}

// GetAllContainers returns all the containers in the database
func GetAllContainers(dbc db.DB) ([]Container, error) {
	var containers []Container
	records, err := dbc.GetAllRecords(tableContainers)
	if err != nil {
		return nil, err
	}
	// transform the records in containers
	for _, serializedContainer := range records {
		c, err := DeserializeContainer(serializedContainer)
		if err != nil {
			return nil, err
		}
		containers = append(containers, c)
	}
	return containers, nil
}

// ContainerUniqueName returns the unique name for the container
func ContainerUniqueName(publicationPoint, campaign string) string {
	return fmt.Sprintf(uniqueContainerNameFormat, publicationPoint, campaign)
}

// DeserializeContainer attempts to convert the string in input in a container object
func DeserializeContainer(c string) (Container, error) {
	var container Container
	err := json.Unmarshal([]byte(c), &container)
	if err != nil {
		return Container{}, err
	}
	return container, nil
}
