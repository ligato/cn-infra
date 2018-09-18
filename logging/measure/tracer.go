// Copyright (c) 2018 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package measure

//go:generate protoc --proto_path=model/apitrace --gogo_out=model/apitrace model/apitrace/apitrace.proto

import (
	"sync"
	"time"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure/model/apitrace"
)

// Tracer allows to measure, store and list measured time entries.
type Tracer interface {
	// LogTime puts measured time to the table and resets the time.
	LogTime(entity string, start time.Time)
	// Get all trace entries stored
	Get() *apitrace.Trace
	// Clear removes entries from the log database
	Clear()
}

// NewTracer creates new tracer object
func NewTracer(name string, log logging.Logger) Tracer {
	return &tracer{
		name:   name,
		log:    log,
		index:  1,
		timedb: make([]*entry, 0),
	}
}

// Inner structure handling database and measure results
type tracer struct {
	sync.Mutex

	name string
	log  logging.Logger
	// Entry index, used in database as key and increased after every entry. Never resets since the tracer object is
	// created or the database is cleared
	index int
	// Time database, uses index as key and entry as value
	timedb []*entry
}

// Single time entry
type entry struct {
	index      int
	name       string
	loggedTime time.Duration
}

func (t *tracer) LogTime(entity string, start time.Time) {
	if t == nil {
		return
	}

	t.Lock()
	defer t.Unlock()

	// Store time
	t.timedb = append(t.timedb, &entry{
		index:      t.index,
		name:       entity,
		loggedTime: time.Since(start),
	})
	t.index++
}

func (t *tracer) Get() *apitrace.Trace {
	t.Lock()
	defer t.Unlock()

	var (
		average map[string][]time.Duration // message name -> measured times
		data    []*apitrace.Trace_TracedEntry
		total   time.Duration
	)

	average = make(map[string][]time.Duration)

	for _, entry := range t.timedb {
		// Add to total
		total += entry.loggedTime
		// Add to message data
		message := &apitrace.Trace_TracedEntry{
			Index:    uint32(entry.index),
			MsgName:  entry.name,
			Duration: entry.loggedTime.String(),
		}
		data = append(data, message)
		// Add to map for average data
		average[entry.name] = append(average[entry.name], entry.loggedTime)
	}

	// Prepare list of average times
	var averageList []*apitrace.Trace_Average
	for msgName, times := range average {
		var total time.Duration
		for _, timeVal := range times {
			total += timeVal
		}
		averageTime := total.Nanoseconds() / int64(len(times))

		averageList = append(averageList, &apitrace.Trace_Average{
			MsgName:     msgName,
			AverageTime: time.Duration(averageTime).String(),
		})
	}

	// Log overall time
	return &apitrace.Trace{
		TracedEntries: data,
		AverageTimes:  averageList,
		Overall:       total.String(),
	}
}

func (t *tracer) Clear() {
	t.timedb = make([]*entry, 0)
}
