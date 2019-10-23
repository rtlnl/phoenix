package logs

import (
	"os"

	"github.com/rs/zerolog"
)

// StdoutLog is the object that handles the writing to stdout
type StdoutLog struct {
	Client zerolog.Logger
}

// NewStdoutLog creates a new object for logging the recommendations to the stdoutput
func NewStdoutLog() StdoutLog {
	return StdoutLog{
		Client: zerolog.New(os.Stdout),
	}
}

func (s StdoutLog) Write(rl RowLog) error {
	for _, itemScore := range rl.ItemScores {
		// create the log message
		msg, err := CreateLogMessage(rl.PublicationPoint, rl.Campaign, rl.SignalID, itemScore)
		if err != nil {
			return err
		}
		// log it on screen
		s.Client.Info().Msg(string(msg))
	}
	return nil
}
