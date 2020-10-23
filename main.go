package main

import (
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	kafka "github.com/bwNetFlow/kafkaconnector"
)

var (
	// common options
	logFile = flag.String("log", "./consumer_dashboard.log", "Location of the log file.")

	// Kafka options
	kafkaConsumerGroup = flag.String("kafka.consumer_group", "dashboard", "Kafka Consumer Group")
	kafkaInTopic       = flag.String("kafka.topic", "flow-messages-enriched", "Kafka topic to consume from")
	kafkaBroker        = flag.String("kafka.brokers", "127.0.0.1:9092,[::1]:9092", "Kafka brokers separated by commas")
	kafkaUser          = flag.String("kafka.user", "", "Kafka username to authenticate with")
	kafkaPass          = flag.String("kafka.pass", "", "Kafka password to authenticate with")
)

// KafkaConn holds the global kafka connection
var kafkaConn = kafka.Connector{}
var promExporter = Exporter{}

func main() {

	flag.Parse()
	if *logFile != "" {
		logfile, err := os.OpenFile(*logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			println("Error opening file for logging: %v", err)
			return
		}
		defer logfile.Close()
		mw := io.MultiWriter(os.Stdout, logfile)
		log.SetOutput(mw)
	}

	// catch termination signal
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-signals
		log.Println("Received exit signal, kthxbye.")
		os.Exit(0)
	}()

	// Enable Prometheus Export
	promExporter.Initialize(":8080")

	// Set kafka auth
	if *kafkaUser != "" {
		kafkaConn.SetAuth(*kafkaUser, *kafkaPass)
	} else {
		kafkaConn.SetAuthAnon()
	}

	// Establish Kafka Connection
	kafkaConn.StartConsumer(*kafkaBroker, []string{*kafkaInTopic}, *kafkaConsumerGroup, -1)
	defer kafkaConn.Close()

	// handle kafka flow messages in foreground
	for {
		promExporter.Increment(<-kafkaConn.ConsumerChannel())
	}
}
