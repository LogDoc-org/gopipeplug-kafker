package main

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/gurkankaymak/hocon"
	"log"
	"strconv"
	"time"
)

type Kafker struct {
	kafkaProducer sarama.SyncProducer
	props         map[string]interface{}
	configured    bool
}

type WatcherMetrics struct {
	entryCountable    bool
	cycleRepeatable   bool
	totalEntryCounter int
	cycleEntryCounter int
	cycleCounter      int
	cycleEntryLimit   int
	cycleLimit        int
	lastMatchedEntry  *LogEntry
	watcherCreated    time.Time
	lastMatched       time.Time
}

type LogEntry struct {
	Ip      string
	SrcTime string
	RcvTime string
	Pid     string
	Source  string
	Entry   string
	AppName string
	Level   int

	fields map[string]Field
}

type Field struct {
	Name  string
	Value string
}

func New() *Kafker {
	return &Kafker{
		configured: false,
	}
}

func (k *Kafker) Configure(config *hocon.Config) error {
	if k.configured || config == nil {
		return nil
	}

	properties := make(map[string]interface{})
	for key, value := range config.GetStringMap("kafka.properties") {
		properties[key] = value
	}

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.Return.Successes = true
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll

	producer, err := sarama.NewSyncProducer(config.GetStringSlice("kafka.brokers"), kafkaConfig)
	if err != nil {
		return err
	}

	k.kafkaProducer = producer

	properties["topic"] = config.GetString("kafka.topic")

	if properties["topic"] == nil {
		return fmt.Errorf("kafka gate is not configured")
	}

	log.Print("Publisher initialized")

	k.configured = true
	k.props = properties

	return nil
}

func (k *Kafker) WatcherStarted(watcherID string) {
	log.Print("Watcher ", watcherID, " started")
}

func (k *Kafker) WatcherCycled(watcherID string) {
	log.Print("Watcher ", watcherID, " cycled")
}

func (k *Kafker) WatcherStopped(watcherID string) {
	log.Print("Watcher ", watcherID, " stopped")
}

func (k *Kafker) Fire(watcherID string, entry *LogEntry, metrics *WatcherMetrics, ctx map[string]string) error {
	if !k.configured {
		return fmt.Errorf("plugin is not configured")
	}

	jsonEntry, err := json.Marshal(entry)
	if err != nil {
		log.Print("Failed to convert log entry to JSON")
		return err
	}

	k.enrichWithCycleRepeatable(entry, metrics)
	k.enrichWithEntryCountable(entry, metrics)

	topic := k.props["topic"].(string)

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(jsonEntry),
	}

	_, _, err = k.kafkaProducer.SendMessage(msg)
	if err != nil {
		log.Print("Failed to send message to Kafka, ", err)
		return err
	}

	return nil
}

func (k *Kafker) enrichWithCycleRepeatable(entry *LogEntry, metrics *WatcherMetrics) {
	if metrics.cycleRepeatable {
		field := Field{
			Name:  "repeats",
			Value: fmt.Sprintf("%d/%d", metrics.cycleCounter, metrics.cycleLimit),
		}
		entry.fields["repeats"] = field

		field = Field{
			Name:  "server_time",
			Value: time.Now().Format(time.RFC3339),
		}
		entry.fields["server_time"] = field
	}
}

func (k *Kafker) enrichWithEntryCountable(entry *LogEntry, metrics *WatcherMetrics) {
	if metrics.entryCountable {
		field := Field{
			Name:  "total",
			Value: strconv.Itoa(metrics.totalEntryCounter),
		}
		entry.fields["events.total"] = field

		field = Field{
			Name:  "current_cycle",
			Value: fmt.Sprintf("%d/%d", metrics.cycleEntryCounter, metrics.cycleEntryLimit),
		}
		entry.fields["events.current_cycle"] = field
	}
}
