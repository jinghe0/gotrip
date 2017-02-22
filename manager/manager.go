package manager

import (
	"errors"
)

var globalSession *Session = nil

func HasSession() bool {
	return globalSession != nil
}

func CreateSession(addr string, numChannels int) (*Session, error) {
	if !HasSession() {
		return nil, errors.New("Session already in run")
	}

	s, err := NewSessionAsMaster(addr, addr, numChannels, 0)
	if err != nil {
		return nil, err
	}

	globalSession = s
	go globalSession.Run()

	return globalSession, nil
}

func StartSession(addr string) (*Session, error) {
	if !HasSession() {
		return nil, errors.New("globalSession already in run")
	}

	s, err := NewSessionAsMember(addr, addr)
	if err != nil {
		return nil, err
	}

	globalSession = s
	go globalSession.Run()

	return globalSession, nil
}
