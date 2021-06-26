package log

import (
	"go.uber.org/zap"

	"github.com/saulmaldonado/agones-mc/internal/config"
)

const (
	prefix = "agones-mc-"
)

func NewLogger(env config.Environment, subcommand config.Subcommand) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	if env == config.Production {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		return nil, err
	}

	return logger.Named(prefix + string(subcommand)), err
}
