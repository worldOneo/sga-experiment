package main

import (
	"sync"
	"time"
)

// MessengerEpoch defines the beginning of the id generation
const MessengerEpoch = 1610468266715

const timeBits = 42
const idBits = 10
const sequenceBits = 11

const maxWorkerID = 1 << idBits
const maxSequence = 1 << sequenceBits

// Snowflake is an int64 used for identification structured like:
//
// -TTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTIIIIIIIIIISSSSSSSSSSS
//
// The fields are:
//   First bit is ignored because of signing
//   (T)ime     - 42bit timestamp in ms
//   (I)D       - 10bit id defined for the current snowflake worker
//   (S)equence - 11bit sequence which keeps incrementing to aviod double ids
type Snowflake int64

// Generator used to generate snowflakes
// holds it's sequence channel
type Generator struct {
	WorkerID     int
	SequenceInt  int
	SequenceSync sync.Mutex
}

//GenSnowflake generates a snowflake
func (sg *Generator) GenSnowflake() Snowflake {
	snowflake := time.Now().UnixNano()
	snowflake /= time.Millisecond.Nanoseconds()
	snowflake -= MessengerEpoch
	snowflake <<= idBits
	snowflake |= int64(sg.WorkerID)
	snowflake <<= sequenceBits
	snowflake |= int64(sg.pushSequence())
	return Snowflake(snowflake)
}

func (sg *Generator) pushSequence() (currentSequence int) {
	sg.SequenceSync.Lock()
	sg.SequenceInt %= maxSequence
	currentSequence = sg.SequenceInt
	sg.SequenceInt++
	sg.SequenceSync.Unlock()
	return
}

// NewGenerator create a snowflake generator
func NewGenerator(workerID int) *Generator {
	return &Generator{
		WorkerID: workerID % maxWorkerID,
	}
}