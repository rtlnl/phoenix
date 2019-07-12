package public

import (
	"fmt"

	"github.com/kataras/iris"
	"github.com/spf13/viper"
)

// Public is the struct that will retain the server for ingesting the
// event from the trackers
type Public struct {
	App      *iris.Application
	Backbone string
}

// NewPublic creates a new Collector object
func NewPublic() (*Public, error) {
	return &Public{
		App: iris.Default(),
	}, nil
}

// Run will initialize the server and will listen to the specified
// port from the config file
func (c *Public) Run() error {
	c.App.Get("/collect", collect)

	host := fmt.Sprintf("%s:%d", viper.GetString("collector.url"), viper.GetInt("collector.port"))
	addr := iris.Addr(host)

	return c.App.Run(addr)
}

// collect will take care of getting the incoming event and do something with it
func collect(ctx iris.Context) {
	ctx.JSON(iris.Map{
		"message": "collected",
	})
}
