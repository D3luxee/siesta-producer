/* Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. */

package producer

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/d3luxee/siesta"
)

func TestProducerSend1(t *testing.T) {
	connector := testConnector(t)
	producerConfig := NewProducerConfig()
	producerConfig.BatchSize = 1

	producer := NewKafkaProducer(producerConfig, ByteSerializer, StringSerializer, connector)
	recordMetadata := producer.Send(&ProducerRecord{Topic: "siesta", Value: "hello world"})

	select {
	case metadata := <-recordMetadata:
		assert(t, metadata.Error, siesta.ErrNoError)
		assert(t, metadata.Topic, "siesta")
		assert(t, metadata.Partition, int32(0))
	case <-time.After(5 * time.Second):
		t.Error("Could not get produce response within 5 seconds")
	}

	producer.Close()
}

func TestProducerSend1000(t *testing.T) {
	connector := testConnector(t)
	producerConfig := NewProducerConfig()
	producerConfig.BatchSize = 100

	producer := NewKafkaProducer(producerConfig, ByteSerializer, StringSerializer, connector)
	metadatas := make([]<-chan *RecordMetadata, 1000)
	for i := 0; i < 1000; i++ {
		metadatas[i] = producer.Send(&ProducerRecord{Topic: "siesta", Value: fmt.Sprintf("%d", i)})
	}

	for _, metadataChan := range metadatas {
		select {
		case metadata := <-metadataChan:
			assert(t, metadata.Error, siesta.ErrNoError)
			assert(t, metadata.Topic, "siesta")
			assert(t, metadata.Partition, int32(0))
		case <-time.After(5 * time.Second):
			t.Fatal("Could not get produce response within 5 seconds")
		}
	}

	producer.Close()
}

func TestProducerRequiredAcks0(t *testing.T) {
	connector := testConnector(t)
	producerConfig := NewProducerConfig()
	producerConfig.BatchSize = 100
	producerConfig.RequiredAcks = 0

	producer := NewKafkaProducer(producerConfig, ByteSerializer, StringSerializer, connector)

	metadatas := make([]<-chan *RecordMetadata, 1000)
	for i := 0; i < 1000; i++ {
		metadatas[i] = producer.Send(&ProducerRecord{Topic: "siesta", Value: fmt.Sprintf("%d", i)})
	}

	for _, metadataChan := range metadatas {
		select {
		case metadata := <-metadataChan:
			assert(t, metadata.Error, siesta.ErrNoError)
			assert(t, metadata.Topic, "siesta")
			assert(t, metadata.Partition, int32(0))
			assert(t, metadata.Offset, int64(-1))
		case <-time.After(5 * time.Second):
			t.Fatal("Could not get produce response within 5 seconds")
		}
	}

	producer.Close()
}

func TestProducerFlushTimeout(t *testing.T) {
	connector := testConnector(t)
	producerConfig := NewProducerConfig()
	producerConfig.RequiredAcks = 0
	producerConfig.Linger = 500 * time.Millisecond

	producer := NewKafkaProducer(producerConfig, ByteSerializer, StringSerializer, connector)
	metadatas := make([]<-chan *RecordMetadata, 100)
	for i := 0; i < 100; i++ {
		metadatas[i] = producer.Send(&ProducerRecord{Topic: "siesta", Value: fmt.Sprintf("%d", i)})
	}

	for _, metadataChan := range metadatas {
		select {
		case metadata := <-metadataChan:
			assert(t, metadata.Error, siesta.ErrNoError)
			assert(t, metadata.Topic, "siesta")
			assert(t, metadata.Partition, int32(0))
			assert(t, metadata.Offset, int64(-1))
		case <-time.After(5 * time.Second):
			t.Fatal("Could not get produce response within 5 seconds")
		}
	}

	producer.Close()
}

func TestProducerWithSeveralTopics(t *testing.T) {
	connector := testConnector(t)
	producerConfig := NewProducerConfig()
	producerConfig.BatchSize = 100
	producerConfig.RequiredAcks = 1

	producer := NewKafkaProducer(producerConfig, ByteSerializer, StringSerializer, connector)
	metadatas := make([]<-chan *RecordMetadata, 1000)
	for i := 0; i < 1000; i++ {
		topic := fmt.Sprintf("siesta-%d", rand.Intn(10))
		metadatas[i] = producer.Send(&ProducerRecord{Topic: topic, Value: fmt.Sprintf("%d", i)})
	}

	for _, metadataChan := range metadatas {
		select {
		case metadata := <-metadataChan:
			assert(t, metadata.Error, siesta.ErrNoError)
		case <-time.After(10 * time.Second):
			t.Fatal("Could not get produce response within 5 seconds")
		}
	}
}
