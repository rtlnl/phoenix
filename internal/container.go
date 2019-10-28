package internal

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rtlnl/phoenix/models"
	"github.com/rtlnl/phoenix/pkg/prisma"
	"github.com/rtlnl/phoenix/utils"
)

// ManagementContainerRequest handles the request from the client
type ManagementContainerRequest struct {
	PublicationPoint string `json:"publicationPoint" binding:"required"`
	Campaign         string `json:"campaign" binding:"required"`
	Model            string `json:"model"`
}

// ManagementContainerResponse handles the response object to the client
type ManagementContainerResponse struct {
	Container *prisma.Container `json:"container"`
	Message   string            `json:"message"`
}

// GetContainer returns an already existsing container
func GetContainer(c *gin.Context) {
	pc := c.MustGet("PrismClient").(*prisma.Client)
	cID := c.Query("id")

	// fetch container
	container, err := models.GetContainerByID(cID, pc)
	if err != nil {
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
	pc := c.MustGet("PrismClient").(*prisma.Client)

	var mc ManagementContainerRequest
	if err := c.BindJSON(&mc); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	container, err := models.NewContainer(mc.PublicationPoint, mc.Campaign, mc.Models, pc)
	if err != nil {
		utils.ResponseError(c, http.StatusUnprocessableEntity, err)
		return
	}

	utils.Response(c, http.StatusCreated, &ManagementContainerResponse{
		Container: container,
		Message:   "container created",
	})
}

// DeleteContainer deletes the container from the database
func DeleteContainer(c *gin.Context) {
	pc := c.MustGet("PrismaClient").(*prisma.Client)
	cID := c.Query("id")

	// empty model from database
	container, err := models.DeleteContainer(cID, pc)
	if err != nil {
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
	pc := c.MustGet("PrismaClient").(*prisma.Client)
	cID := c.Query("id")

	var mc ManagementContainerRequest
	if err := c.BindJSON(&mc); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	// check if model's value is empty
	if mc.Model == "" {
		utils.ResponseError(c, http.StatusBadRequest, errors.New("model parameter is empty"))
		return
	}

	// link the models internally
	container, err := models.LinkModel(cID, mc.Model, pc)
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
	Containers []prisma.Container `json:"containers"`
	Message    string             `json:"message"`
}

// GetAllContainers returns all the containers in the database
func GetAllContainers(c *gin.Context) {
	pc := c.MustGet("PrismClient").(*prisma.Client)

	// fetch container
	containers, err := models.GetAllContainers(pc)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}

	utils.Response(c, http.StatusOK, &ManagementContainersResponse{
		Containers: containers,
		Message:    "containers fetched",
	})
}
