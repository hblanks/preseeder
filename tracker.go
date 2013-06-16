package preseeder

import (
	"time"
)

type LogEvent struct {
	Name string    `json:"name"`
	Time time.Time `json:"time"`
}

type ClientRecord struct {
	Addr       string     `json:"address"`
	MacAddress string     `json:"mac"`
	Events     []LogEvent `json:"events"`
}

type TrackerState map[string]*ClientRecord

type ClientTracker struct {
	state      TrackerState
	writeEvent chan *trackerEvent
	readState  chan TrackerState
	exit       chan int
	terminated chan int
}

type trackerEvent struct {
	name       string
	addr       string
	macAddress string
}

func NewClientTracker() *ClientTracker {
	return &ClientTracker{
		writeEvent: make(chan *trackerEvent),
		readState:  make(chan TrackerState),
		state:      make(TrackerState),
		exit:       make(chan int),
		terminated: make(chan int),
	}
}

func (ct *ClientTracker) run() {
	var event *trackerEvent
	var record *ClientRecord
	defer func() { ct.terminated <- 1 }()
	for {
		select {
		case event = <-ct.writeEvent:
			record = ct.state[event.addr]
			if ct.state[event.addr] == nil {
				record = &ClientRecord{
					Addr:       event.addr,
					MacAddress: event.macAddress,
				}
				ct.state[event.addr] = record
			}
			if record.MacAddress == "" && event.macAddress != "" {
				record.MacAddress = event.macAddress
			}
			record.Events = append(record.Events,
				LogEvent{event.name, time.Now()})

		case ct.readState <- ct.state:
			continue

		case <-ct.exit:
			return
		}
	}
}

func (ct *ClientTracker) Track(name, addr string, macAddress string) {
	ct.writeEvent <- &trackerEvent{name, addr, macAddress}
}

func (ct *ClientTracker) ReadState() TrackerState {
	return <-ct.readState
}

func (ct *ClientTracker) Start() {
	go ct.run()
}

func (ct *ClientTracker) Stop() {
	ct.exit <- 1
	<-ct.terminated
}
