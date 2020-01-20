package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewContainer(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	container, err := NewContainer("publication", "campaign", nil, dbc)
	if err != nil {
		t.FailNow()
	}

	assert.Equal(t, "publication", container.PublicationPoint)
	assert.Equal(t, "campaign", container.Campaign)
	assert.Equal(t, 0, len(container.Models))
}

func TestNewContainerWithModel(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	if _, err := NewModel("model", "", []string{"signal"}, dbc); err != nil {
		t.FailNow()
	}

	container, err := NewContainer("publication1", "campaign", []string{"model"}, dbc)
	if err != nil {
		t.FailNow()
	}

	assert.Equal(t, "publication1", container.PublicationPoint)
	assert.Equal(t, "campaign", container.Campaign)
	assert.Equal(t, 1, len(container.Models))
}

func TestContainerExists(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	exists := ContainerExists("publication", "campaign", dbc)

	assert.Equal(t, true, exists)
}

func TestGetContainer(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	container, err := GetContainer("publication1", "campaign", dbc)
	if err != nil {
		t.FailNow()
	}

	assert.Equal(t, "publication1", container.PublicationPoint)
	assert.Equal(t, "campaign", container.Campaign)
	assert.Equal(t, 1, len(container.Models))
}

func TestDeleteContainer(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	container, err := GetContainer("publication1", "campaign", dbc)
	if err != nil {
		t.FailNow()
	}
	err = container.DeleteContainer(dbc)
	if err != nil {
		t.FailNow()
	}
}

func TestGetAllContainers(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	_, err := NewContainer("publication3", "campaign", nil, dbc)
	if err != nil {
		t.FailNow()
	}

	containers, count, err := GetAllContainers(dbc)
	if err != nil {
		t.FailNow()
	}

	assert.Equal(t, 2, len(containers))
	assert.Equal(t, 2, count)
}

func TestContainerUniqueName(t *testing.T) {
	un := ContainerUniqueName("hello", "world")
	assert.Equal(t, "hello:world", un)
}

func TestDeserializeContainer(t *testing.T) {
	ser := `{"publicationPoint":"pub","campaign":"cmp","models":["model"]}`
	c, err := DeserializeContainer(ser)
	if err != nil {
		t.FailNow()
	}
	assert.Equal(t, "pub", c.PublicationPoint)
	assert.Equal(t, "cmp", c.Campaign)
	assert.Equal(t, 1, len(c.Models))
}

func TestLinkModel(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	container, err := GetContainer("publication", "campaign", dbc)
	if err != nil {
		t.FailNow()
	}

	if _, err := NewModel("model1", "", []string{"signal"}, dbc); err != nil {
		t.FailNow()
	}

	if _, err := NewModel("model2", "", []string{"signal"}, dbc); err != nil {
		t.FailNow()
	}

	err = container.LinkModel([]string{"model1", "model2"}, dbc)
	if err != nil {
		t.FailNow()
	}

	assert.Equal(t, 2, len(container.Models))
}
