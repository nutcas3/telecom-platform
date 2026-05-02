package id

import (
	"fmt"
	"sync"
	"time"
)

const (
	Epoch int64 = 1577836800000

	NodeIDBits = 5
	SequenceBits = 12

	MaxNodeID = -1 ^ (-1 << NodeIDBits)
	MaxSequence = -1 ^ (-1 << SequenceBits)
	NodeShift = SequenceBits
	TimeShift = SequenceBits + NodeIDBits
)

type Snowflake struct {
	mu        sync.Mutex
	timestamp int64
	nodeID    int64
	sequence  int64
}

func NewSnowflake(nodeID int64) (*Snowflake, error) {
	if nodeID < 0 || nodeID > MaxNodeID {
		return nil, fmt.Errorf("node ID must be between 0 and %d", MaxNodeID)
	}

	return &Snowflake{
		nodeID: nodeID,
	}, nil
}

func (s *Snowflake) Generate() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixNano() / 1e6 // Convert to milliseconds

	if s.timestamp == now {
		s.sequence = (s.sequence + 1) & MaxSequence
		if s.sequence == 0 {
			// Sequence overflow, wait for next millisecond
			for now <= s.timestamp {
				now = time.Now().UnixNano() / 1e6
			}
		}
	} else {
		s.sequence = 0
	}

	s.timestamp = now

	return ((now - Epoch) << TimeShift) |
		(s.nodeID << NodeShift) |
		s.sequence
}

func (s *Snowflake) GenerateString() string {
	return fmt.Sprintf("%d", s.Generate())
}

func Parse(id int64) (timestamp, nodeID, sequence int64) {
	timestamp = (id >> TimeShift) + Epoch
	nodeID = (id >> NodeShift) & MaxNodeID
	sequence = id & MaxSequence
	return
}

func Time(id int64) time.Time {
	timestamp, _, _ := Parse(id)
	return time.Unix(timestamp/1000, (timestamp%1000)*1e6)
}

func NodeID(id int64) int64 {
	_, nodeID, _ := Parse(id)
	return nodeID
}

func Sequence(id int64) int64 {
	_, _, sequence := Parse(id)
	return sequence
}

var defaultGenerator *Snowflake

func init() {
	var err error
	defaultGenerator, err = NewSnowflake(1)
	if err != nil {
		panic(err)
	}
}

func Generate() int64 {
	return defaultGenerator.Generate()
}

func GenerateString() string {
	return defaultGenerator.GenerateString()
}

func GeneratePrefixed(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, GenerateString())
}
