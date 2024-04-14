package logger

import "go.uber.org/zap"

type Logger struct {
	Log zap.SugaredLogger
}

func New() (*Logger, error) {

	// Create installed registrator zap
	log, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	defer log.Sync()

	// Create registrator SugaredLogger
	sugar := *log.Sugar()

	return &Logger{Log: sugar}, nil
}
