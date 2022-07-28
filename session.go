package telemetry

import "github.com/google/uuid"

var InvalidSession = Session(uuid.UUID{})

type Session uuid.UUID

func RandomSession() Session {
	return Session(uuid.New())
}

func ParseSession(sess string) (Session, error) {
	s, err := uuid.Parse(sess)
	if err != nil {
		return Session{}, err
	}
	return Session(s), nil
}

func (s Session) String() string {
	return uuid.UUID(s).String()
}
