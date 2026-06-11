package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// roomCodeAlphabet excludes visually ambiguous characters (0/O, 1/I/L) so a
// roomCode read off a slide or whiteboard is easy to type correctly.
const roomCodeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

const roomCodeLength = 6

// newRoomCode returns a random, human-friendly room code drawn uniformly from
// roomCodeAlphabet. Uniqueness is guaranteed by the DB unique constraint plus
// retry on collision in the service, not by this function.
func newRoomCode() (string, error) {
	b := make([]byte, roomCodeLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("service: read random: %w", err)
	}
	out := make([]byte, roomCodeLength)
	for i, v := range b {
		out[i] = roomCodeAlphabet[int(v)%len(roomCodeAlphabet)]
	}
	return string(out), nil
}

// newToken returns an unpredictable token of the form "<prefix>_<48 hex chars>"
// (24 bytes of crypto/rand entropy). Used for teacherToken ("teacher") and
// clientToken ("student").
func newToken(prefix string) (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("service: read random: %w", err)
	}
	return prefix + "_" + hex.EncodeToString(b), nil
}
