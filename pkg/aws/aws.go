package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// NewAWSSession returns a sessions for connecting to the aws services
func NewAWSSession(region, endpoint string, disableSSL bool) *session.Session {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:           aws.String(region),
		S3ForcePathStyle: aws.Bool(true),
		DisableSSL:       aws.Bool(disableSSL),
		Endpoint:         aws.String(endpoint),
	}))

	return sess
}
