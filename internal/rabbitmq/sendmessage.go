package rabbitmq

import "errors"

type SendMessage struct {
	Queue string
	Body  string
}

func (s *SendMessage) Validate() error {
	if s.Queue == "" {
		return errors.New("missing queue")
	}

	if s.Body == "" {
		return errors.New("missing body")
	}

	return nil
}
