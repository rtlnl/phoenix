package tucson

import (
	context "context"
	"time"

	"google.golang.org/grpc"

	"github.com/rs/zerolog/log"
)

// Client holds the gRPC information for connecting to Tucson
type Client struct {
	Conn ServerClient
}

// NewClient creates a new client to connect to the gRPC Tucson API
func NewClient(address string) *Client {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Panic().Msgf("did not connect: %v", err)
	}
	return &Client{
		Conn: NewServerClient(conn),
	}
}

// Ping tests theserver connection
func (c *Client) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := c.Conn.Ping(ctx, &PingMessage{Msg: "Ping"})
	return err
}

// GetModel returns the name of the model based on the publicationPoint and campaign in input
func (c *Client) GetModel(publicationPoint, campaign string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.Conn.GetModel(ctx, &ModelRequestMessage{
		PublicationPoint: publicationPoint,
		Campaign:         campaign,
	})
	if err != nil {
		return "", err
	}

	return r.GetName(), nil
}
