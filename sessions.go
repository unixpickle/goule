package goule

import (
	"math/rand"
	"strconv"
	"time"
)

type sessions struct {
	goule       *Goule
	startTimes  map[string]time.Time
	lastCleanup time.Time
}

func newSessions(g *Goule) *sessions {
	return &sessions{g, map[string]time.Time{}, time.Now()}
}

func (s *sessions) cleanup() {
	remaining := make(map[string]time.Time)
	for key, start := range s.startTimes {
		if time.Since(start) < s.timeout() {
			remaining[key] = start
		}
	}
	s.startTimes = remaining
	s.lastCleanup = time.Now()
}

func (s *sessions) conditionalCleanup() {
	if time.Since(s.lastCleanup) > s.timeout() {
		s.cleanup()
	}
}

func (s *sessions) login() string {
	str := strconv.Itoa(rand.Int()) + s.goule.config.Admin.Hash
	hash := Hash(str)
	s.startTimes[hash] = time.Now()
	return hash
}

func (s *sessions) logout(key string) {
	delete(s.startTimes, key)
}

func (s *sessions) timeout() time.Duration {
	return time.Second * time.Duration(s.goule.config.Admin.Timeout)
}

func (s *sessions) validate(key string) bool {
	s.conditionalCleanup()
	start, ok := s.startTimes[key]
	if !ok {
		return false
	}
	if time.Since(start) > s.timeout() {
		delete(s.startTimes, key)
		return false
	}
	return true
}
