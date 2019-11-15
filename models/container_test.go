package models

import "testing"

import "github.com/stretchr/testify/assert"

func TestNewContainer(t *testing.T) {
	ac, close := GetTestAerospikeClient()
	defer close()

	c, err := NewContainer("publication", "campaign", nil, ac)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, "publication", c.PublicationPoint)
	assert.Equal(t, "campaign", c.Campaign)
	assert.Equal(t, 0, len(c.Models))
}

func TestNewContainerWithModel(t *testing.T) {
	ac, close := GetTestAerospikeClient()
	defer close()

	if m, err := NewModel("model", "", []string{"signal"}, ac); m == nil || err != nil {
		t.Fail()
	}

	c, err := NewContainer("publication", "campaign", []string{"model"}, ac)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, "publication", c.PublicationPoint)
	assert.Equal(t, "campaign", c.Campaign)
	assert.Equal(t, 1, len(c.Models))
}
