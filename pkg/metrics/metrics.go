package metrics

// Metrics is the interface that will be responsible for collecting metrics
type Metrics interface {
	// FailedRequest keeps track of those requests with both status code 400 and 500
	FailedRequest()
	// SuccessRequest keeps track of the requests with status code 200
	SuccessRequest()
	// NotFoundRequest keeps track of the requests with status code 404
	NotFoundRequest()
	// StartTimer will initialize the timer for calculating the latency
	StartTimer()
	// Latency measure the latency from when the request hits the endpoint to the response
	Latency()
}
