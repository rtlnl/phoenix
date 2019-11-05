package internal

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rtlnl/phoenix/models"
	"github.com/rtlnl/phoenix/pkg/db"
	"github.com/rtlnl/phoenix/utils"
)

// ManagementContainerRequest handles the request from the client
type ManagementContainerRequest struct {
	PublicationPoint string   `json:"publicationPoint" binding:"required"`
	Campaign         string   `json:"campaign" binding:"required"`
	Models           []string `json:"models"`
}

// ManagementContainerResponse handles the response object to the client
type ManagementContainerResponse struct {
	Container *models.Container `json:"container"`
	Message   string            `json:"message"`
}

// GetContainer returns an already existsing container
func GetContainer(c *gin.Context) {
	ac := c.MustGet("AerospikeClient").(*db.AerospikeClient)

	// read from params in url
	pp := c.Query("publicationPoint")
	cmp := c.Query("campaign")

	// if either is empty then
	if pp == "" || cmp == "" {
		utils.ResponseError(c, http.StatusBadRequest, errors.New("missing parameters in url for searching the container"))
		return
	}

	// fetch container
	container, err := models.GetExistingContainer(pp, cmp, ac)
	if container == nil || err != nil {
		utils.ResponseError(c, http.StatusNotFound, fmt.Errorf("container with publication point %s and campaign %s not found", pp, cmp))
		return
	}

	utils.Response(c, http.StatusOK, &ManagementContainerResponse{
		Container: container,
		Message:   "container fetched",
	})
}

// CreateContainer creates a new container for the given publication point and campaign
func CreateContainer(c *gin.Context) {
	ac := c.MustGet("AerospikeClient").(*db.AerospikeClient)

	var mc ManagementContainerRequest
	if err := c.BindJSON(&mc); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	container, err := models.NewContainer(mc.PublicationPoint, mc.Campaign, mc.Models, ac)
	if err != nil {
		utils.ResponseError(c, http.StatusUnprocessableEntity, err)
		return
	}

	utils.Response(c, http.StatusCreated, &ManagementContainerResponse{
		Container: container,
		Message:   "container created",
	})
}

// EmptyContainer truncate the container's data
func EmptyContainer(c *gin.Context) {
	ac := c.MustGet("AerospikeClient").(*db.AerospikeClient)

	var mc ManagementContainerRequest
	if err := c.BindJSON(&mc); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	container, err := models.GetExistingContainer(mc.PublicationPoint, mc.Campaign, ac)
	if container == nil || err != nil {
		utils.ResponseError(c, http.StatusNotFound, fmt.Errorf("container with publication point %s and campaign %s not found", mc.PublicationPoint, mc.Campaign))
		return
	}

	// empty model from database
	if err := container.DeleteContainer(ac); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	utils.Response(c, http.StatusOK, &ManagementContainerResponse{
		Container: container,
		Message:   "container empty",
	})
}

// LinkModel attaches the specified models in input to an existing container
func LinkModel(c *gin.Context) {
	ac := c.MustGet("AerospikeClient").(*db.AerospikeClient)

	var mc ManagementContainerRequest
	if err := c.BindJSON(&mc); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	if len(mc.Models) <= 0 {
		utils.ResponseError(c, http.StatusBadRequest, errors.New("no models to link to the container"))
		return
	}

	// get the existing model
	container, err := models.GetExistingContainer(mc.PublicationPoint, mc.Campaign, ac)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, fmt.Errorf("container with publication point %s and campaign %s not found", mc.PublicationPoint, mc.Campaign))
		return
	}

	// link the models internally
	container, err = container.LinkModel(mc.Models, ac)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	utils.Response(c, http.StatusOK, &ManagementContainerResponse{
		Container: container,
		Message:   "model linked to container",
	})
}

// ManagementContainersResponse handles the response when there are multiple containers
type ManagementContainersResponse struct {
	Containers []*models.Container `json:"containers"`
	Message    string              `json:"message"`
}

// GetAllContainers returns all the containers in the database
func GetAllContainers(c *gin.Context) {
	ac := c.MustGet("AerospikeClient").(*db.AerospikeClient)

	// fetch container
	containers, err := models.GetAllContainers(ac)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, errors.New("no containers found"))
		return
	}

	utils.Response(c, http.StatusOK, &ManagementContainersResponse{
		Containers: containers,
		Message:    "containers fetched",
	})
}
