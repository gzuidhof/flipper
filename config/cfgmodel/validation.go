package cfgmodel

import (
	"errors"
	"time"
)

func checkDuration(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return errors.New("must be a string")
	}

	if s == "" {
		return nil
	}

	_, err := time.ParseDuration(s)
	if err != nil {
		return errors.New("invalid duration")
	}
	return nil
}
