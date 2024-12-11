package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"time"

	sysinfo "github.com/elastic/go-sysinfo"
	cron "github.com/go-co-op/gocron/v2"
	amqp "github.com/rabbitmq/amqp091-go"
	yaml "gopkg.in/yaml.v3"
)

type AMQPConnection struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
}

// Can reuse for updates on the machine status
// Possible library https://pkg.go.dev/github.com/zcalusic/sysinfo
type MemoryMetric struct {
	TotalMemory     uint64
	AvailableMemory uint64
}

// TODO Get some metrics on cpu
type SystemInfo struct {
	Hostname string
	Ip       []string
	Mac      []string
	BootTime time.Time
	Memory   MemoryMetric
}

const statusQueueName = "status"

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	amqpConn := getAMQPConnection()
	conn, err := amqp.Dial(amqpConn)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	activateAgent := true

	switch environment := getExecutionEnvironment(); environment {
	case "Containerised":
		log.Print("Agent deactivated as application is already managed by K8s")
		activateAgent = false
	case "Non Containerised":
		log.Print("Agent activated to bring application and device under K8s management")
	default:
		log.Print(environment)
		activateAgent = false
	}

	if !activateAgent {
		return
	}

	registerDevice(ch)
	scheduleHeartbeats(ch)

	switchQueue, err := ch.QueueDeclare(
		"switchImage", // name
		false,         // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	failOnError(err, "Failed to declare a queue")

	switchMsgs, err := ch.Consume(
		switchQueue.Name, // queue
		"",               // consumer
		true,             // auto-ack
		false,            // exclusive
		false,            // no-local
		false,            // no-wait
		nil,              // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for msg := range switchMsgs {
			if msg.ContentType == "application/x-sh" {
				image := string(msg.Body)
				log.Printf("Attempting to reboot system with %s based bootc image", image)

				switchCmd := exec.Command("bootc", "switch", image)
				stderrPipe, err := switchCmd.StderrPipe()

				if err != nil {
					log.Print(err)
				}

				if err := switchCmd.Start(); err != nil {
					log.Print(err)
				}

				if stderr, _ := io.ReadAll(stderrPipe); len(stderr) > 0 {
					log.Print(string(stderr))
					continue
				}

				log.Print("Switch successful, system will be rebooted")
				rebootCmd := exec.Command("reboot")
				rebootCmd.Start()
			} else {
				log.Printf("Received message: %s", msg.Body)
			}
		}
	}()

	<-forever
}

func getAMQPConnection() string {
	yamlFile, err := os.ReadFile("/etc/amqp-config.yaml")
	failOnError(err, "Failed to read AMQP config file")

	var amqpConn AMQPConnection
	err = yaml.Unmarshal(yamlFile, &amqpConn)
	failOnError(err, "Failed to unmarshal AMQP config file")

	user := amqpConn.User
	password := amqpConn.Password
	host := amqpConn.Host
	port := amqpConn.Port
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", user, password, host, port)
}

func registerDevice(ch *amqp.Channel) {
	hostSys, err := sysinfo.Host()
	failOnError(err, "Failed to get host system information")
	hostSysMemoryInfo, err := hostSys.Memory()
	failOnError(err, "Failed to get system memory information")

	systemInfo := SystemInfo{
		Hostname: hostSys.Info().Hostname,
		Ip:       hostSys.Info().IPs,
		Mac:      hostSys.Info().MACs,
		BootTime: hostSys.Info().BootTime,
		Memory: MemoryMetric{
			TotalMemory:     hostSysMemoryInfo.Total,
			AvailableMemory: hostSysMemoryInfo.Available,
		},
	}

	registerQueue, err := ch.QueueDeclare(
		"register", // name
		false,      // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	failOnError(err, "Failed to declare a queue")

	body, err := json.Marshal(systemInfo)
	failOnError(err, "Failed to marshal")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx,
		"",                 // exchange
		registerQueue.Name, // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	failOnError(err, "Failed to publish register message")
}

func scheduleHeartbeats(ch *amqp.Channel) {
	_, err := ch.QueueDeclare(
		statusQueueName, // name
		false,           // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	failOnError(err, "Failed to declare a queue")

	scheduler, err := cron.NewScheduler()
	failOnError(err, "Failed to create heartbeat scheduler")

	_, err = scheduler.NewJob(cron.DurationJob(60*time.Second), cron.NewTask(relayHeartbeat, ch))
	failOnError(err, "Failed to create heartbeat job")

	scheduler.Start()
}

func relayHeartbeat(ch *amqp.Channel) {
	status, err := exec.Command("bootc", "status").Output()

	failOnError(err, "Failed to call bootc status")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx,
		"",              // exchange
		statusQueueName, // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        status,
		})
	failOnError(err, "Failed to publish heartbeat message")
}

func getExecutionEnvironment() string {
	content, osErr := os.ReadFile("/proc/1/environ")
	if osErr != nil {
		return fmt.Sprintf("Undetermined execution environment: %s", osErr)
	}

	match, regexErr := regexp.Match("KUBERNETES", content)
	if regexErr != nil {
		return fmt.Sprintf("Undetermined execution environment: %s", osErr)
	}

	if match {
		return "Containerised"
	} else {
		return "Non Containerised"
	}
}
