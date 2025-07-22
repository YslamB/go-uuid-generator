package snowflake

import (
	"fmt"
	"sync"
	"time"
)

const (
	TwitterEpoch = 1288834974657

	DatacenterIdBits = 5
	MachineIdBits    = 5
	SequenceBits     = 12

	MaxDatacenterId = (1 << DatacenterIdBits) - 1
	MaxMachineId    = (1 << MachineIdBits) - 1
	MaxSequence     = (1 << SequenceBits) - 1

	MachineIdShift    = SequenceBits
	DatacenterIdShift = SequenceBits + MachineIdBits
	TimestampShift    = SequenceBits + MachineIdBits + DatacenterIdBits
)

type Generator struct {
	mu           sync.Mutex
	datacenterId int64
	machineId    int64
	sequence     int64
	lastTime     int64
}

func NewGenerator(datacenterId, machineId int64) (*Generator, error) {

	if datacenterId < 0 || datacenterId > MaxDatacenterId {
		return nil, fmt.Errorf("datacenter ID must be between 0 and %d", MaxDatacenterId)
	}
	if machineId < 0 || machineId > MaxMachineId {
		return nil, fmt.Errorf("machine ID must be between 0 and %d", MaxMachineId)
	}

	return &Generator{
		datacenterId: datacenterId,
		machineId:    machineId,
		sequence:     0,
		lastTime:     -1,
	}, nil
}

func (g *Generator) NextID() (int64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	timestamp := g.timeGen()

	if timestamp < g.lastTime {
		return 0, fmt.Errorf("clock moved backwards. Refusing to generate ID for %d milliseconds", g.lastTime-timestamp)
	}

	if timestamp == g.lastTime {
		g.sequence = (g.sequence + 1) & MaxSequence

		if g.sequence == 0 {
			timestamp = g.tilNextMillis(g.lastTime)
		}

	} else {
		g.sequence = 0
	}

	g.lastTime = timestamp

	id := ((timestamp - TwitterEpoch) << TimestampShift) |
		(g.datacenterId << DatacenterIdShift) |
		(g.machineId << MachineIdShift) |
		g.sequence
	return id, nil
}

func (sg *Generator) timeGen() int64 {
	return time.Now().UnixNano() / 1e6
}

func (sg *Generator) tilNextMillis(lastTimestamp int64) int64 {
	timestamp := sg.timeGen()
	for timestamp <= lastTimestamp {
		timestamp = sg.timeGen()
	}
	return timestamp
}

func (sg *Generator) ParseID(id int64) map[string]interface{} {
	timestamp := (id >> TimestampShift) + TwitterEpoch
	datacenterId := (id >> DatacenterIdShift) & ((1 << DatacenterIdBits) - 1)
	machineId := (id >> MachineIdShift) & ((1 << MachineIdBits) - 1)
	sequence := id & ((1 << SequenceBits) - 1)

	return map[string]interface{}{
		"id":            id,
		"timestamp":     timestamp,
		"datetime":      time.Unix(timestamp/1000, (timestamp%1000)*1e6).UTC().Format(time.RFC3339),
		"datacenter_id": datacenterId,
		"machine_id":    machineId,
		"sequence":      sequence,
	}
}
