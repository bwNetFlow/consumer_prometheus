package main

import (
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwNetFlow/consumer_prometheus/exporter"
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
	kafkaAuthAnon      = flag.Bool("kafka.auth_anon", false, "Set Kafka Auth Anon")
	kafkaDisableTLS    = flag.Bool("kafka.disable_tls", false, "Whether to use tls or not")
	kafkaDisableAuth   = flag.Bool("kafka.disable_auth", false, "Whether to use auth or not")
)

// KafkaConn holds the global kafka connection
var kafkaConn = kafka.Connector{}
var promExporter = exporter.Exporter{}

func main() {
	flag.Parse()
	var err error

	// initialize logger
	logfile, err := os.OpenFile(*logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		println("Error opening file for logging: %v", err)
		return
	}
	defer logfile.Close()
	mw := io.MultiWriter(os.Stdout, logfile)
	log.SetOutput(mw)
	log.Println("-------------------------- Started.")

	// catch termination signal
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-signals
		log.Println("Received exit signal, kthxbye.")
		os.Exit(0)
	}()

	// Enable Prometheus Export
	promExporter.Initialize()
	promExporter.ServeEndpoints(":8080")

	// disable TLS if requested
	if *kafkaDisableTLS {
		log.Println("kafkaDisableTLS ...")
		kafkaConn.DisableTLS()
	}
	if *kafkaDisableAuth {
		log.Println("kafkaDisableAuth ...")
		kafkaConn.DisableAuth()
	} else { // set Kafka auth
		if *kafkaAuthAnon {
			kafkaConn.SetAuthAnon()
		} else if *kafkaUser != "" {
			kafkaConn.SetAuth(*kafkaUser, *kafkaPass)
		} else {
			log.Println("No explicit credentials available, trying env.")
			err = kafkaConn.SetAuthFromEnv()
			if err != nil {
				log.Println("No credentials available, using 'anon:anon'.")
				kafkaConn.SetAuthAnon()
			}
		}
	}

	// Establish Kafka Connection
	err = kafkaConn.StartConsumer(*kafkaBroker, []string{*kafkaInTopic}, *kafkaConsumerGroup, -1)
	if err != nil {
		log.Println("StartConsumer:", err)
		// sleep to make auto restart not too fast and spamming connection retries
		time.Sleep(5 * time.Second)
		return
	}
	defer kafkaConn.Close()

	// handle kafka flow messages in foreground
	for flow := range kafkaConn.ConsumerChannel() {
		promExporter.Increment(flow)
	}
}
