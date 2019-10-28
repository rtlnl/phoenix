package models

import (
	"context"
	"fmt"

	"github.com/rtlnl/phoenix/pkg/prisma"
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
func NewContainer(publicationPoint, campaign string, models []string, pc *prisma.Client) (*prisma.Container, error) {
	// does container exists already then return it to the client
	c, err := GetContainerByParams(publicationPoint, campaign, pc)
	if err != nil {
		return nil, err
	}

	// container already exists
	if c != nil {
		return nil, fmt.Errorf("container with publicationPoint %s and campaign %s already exists", publicationPoint, campaign)
	}

	// search for models if any
	var mods []prisma.ModelWhereUniqueInput
	for _, m := range models {
		mod, err := pc.Model(prisma.ModelWhereUniqueInput{
			Name: &m,
		}).Exec(context.Background())
		if err != nil {
			return nil, err
		}
		// throw an error if model doesn't exist
		if mod == nil {
			return nil, fmt.Errorf("model with name %s not found", m)
		}
		// append models' ID
		mods = append(mods, prisma.ModelWhereUniqueInput{
			ID: &mod.ID,
		})
	}

	// create new container
	container, err := pc.CreateContainer(prisma.ContainerCreateInput{
		PublicationPoint: publicationPoint,
		Campaign:         campaign,
		Models: &prisma.ModelCreateManyWithoutContainerIdInput{
			Connect: mods,
		},
	}).Exec(context.Background())
	if err != nil {
		return nil, err
	}

	return container, nil
}

// GetContainerByParams looks to find a container by publication point and campaign
func GetContainerByParams(publicationPoint, campaign string, pc *prisma.Client) (*prisma.Container, error) {
	c, err := pc.Container(prisma.ContainerWhereUniqueInput{
		PublicationPoint: prisma.Str(publicationPoint),
		Campaign:         prisma.Str(campaign),
	}).Exec(context.Background())

	if err != nil {
		return nil, err
	}
	return c, nil
}

// GetContainerByID looks to find a container by its ID
func GetContainerByID(id string, pc *prisma.Client) (*prisma.Container, error) {
	c, err := pc.Container(prisma.ContainerWhereUniqueInput{
		ID: utils.ConvertStringToInt32(id),
	}).Exec(context.Background())

	if err != nil {
		return nil, err
	}
	return c, nil
}

// DeleteContainer deletes the content of the container by truncating the PublicationPoint (aka setName)
func DeleteContainer(ID string, pc *prisma.Client) (*prisma.Container, error) {
	container, err := pc.DeleteContainer(prisma.ContainerWhereUniqueInput{
		ID: utils.ConvertStringToInt32(ID),
	}).Exec(context.Background())
	return container, err
}

// LinkModel appends the models inside Aerospike structure
func LinkModel(ID string, model string, pc *prisma.Client) (*prisma.Container, error) {
	// search for models if any
	var mods []prisma.ModelWhereUniqueInput
	mod, err := pc.Model(prisma.ModelWhereUniqueInput{
		Name: prisma.Str(model),
	}).Exec(context.Background())
	if err != nil {
		return nil, err
	}
	// throw an error if model doesn't exist
	if mod == nil {
		return nil, fmt.Errorf("model with name %s not found", model)
	}
	// append models' ID
	mods = append(mods, prisma.ModelWhereUniqueInput{
		ID: &mod.ID,
	})

	pc.UpdateContainer(prisma.ContainerUpdateParams{
		Where: prisma.ContainerWhereUniqueInput{
			ID: utils.ConvertStringToInt32(ID),
		},
		Data: prisma.ContainerUpdateInput{
			Models: &prisma.ModelUpdateManyWithoutContainerIdInput{
				Connect: mods,
			},
		},
	})

	return nil, nil
}

// GetAllContainers returns all the containers in the database
func GetAllContainers(pc *prisma.Client) ([]prisma.Container, error) {
	containers, err := pc.Containers(&prisma.ContainersParams{}).Exec(context.Background())
	return containers, err
}
