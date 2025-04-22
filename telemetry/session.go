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

func (s *Session) MarshalJSON() ([]byte, error) {
	return []byte(`"` + s.String() + `"`), nil
}

func (s *Session) UnmarshalJSON(b []byte) error {
	if len(b) < 2 {
		return nil
	}
	b = b[1 : len(b)-1]
	var err error
	*s, err = ParseSession(string(b))
	return err
}
