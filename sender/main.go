package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/streadway/amqp"
)

func main() {
	// Define RabbitMQ server URL.
	amqpServerURL := os.Getenv("AMQP_SERVER_URL")

	// Create a new RabbitMQ connection.
	connectRabbitMQ, err := amqp.Dial(amqpServerURL)
	if err != nil {
		panic(err)
	}
	defer connectRabbitMQ.Close()

	// Let's start by opening a channel to our RabbitMQ
	// instance over the connection we have already
	// established.
	channelRabbitMQ, err := connectRabbitMQ.Channel()
	if err != nil {
		panic(err)
	}
	defer channelRabbitMQ.Close()

	args := amqp.Table{
		"x-message-ttl":             int32(10 * 1000), // 10 seconds
		"x-dead-letter-exchange":    "main.webhook.exchange",
		"x-dead-letter-routing-key": "",
	}

	// With the instance and declare Queues that we can
	// publish and subscribe to.
	err = channelRabbitMQ.ExchangeDeclare(
		"main.webhook.exchange", // name
		"direct",                // type
		true,                    // durable
		false,                   // auto-deleted
		false,                   // internal
		false,                   // no-wait
		nil,                     // arguments
	)

	err = channelRabbitMQ.ExchangeDeclare(
		"retry.webhook.exchange", // name
		"direct",                 // type
		true,                     // durable
		false,                    // auto-deleted
		false,                    // internal
		false,                    // no-wait
		nil,                      // arguments
	)

	_, err = channelRabbitMQ.QueueDeclare(
		"main.webhook.queue", // queue name
		true,                 // durable
		false,                // auto delete
		false,                // exclusive
		false,                // no wait
		nil,                  // arguments
	)
	if err != nil {
		panic(err)
	}

	_ = channelRabbitMQ.QueueBind(
		"main.webhook.queue",    // queue name
		"",                      // routing key
		"main.webhook.exchange", // exchange
		false,
		nil)

	_, err = channelRabbitMQ.QueueDeclare(
		"retry.webhook.queue-1", // queue name
		true,                    // durable
		false,                   // auto delete
		false,                   // exclusive
		false,                   // no wait
		args,                    // arguments
	)
	if err != nil {
		panic(err)
	}

	_, err = channelRabbitMQ.QueueDeclare(
		"retry.webhook.queue-2", // queue name
		true,                    // durable
		false,                   // auto delete
		false,                   // exclusive
		false,                   // no wait
		args,                    // arguments
	)
	if err != nil {
		panic(err)
	}

	_, err = channelRabbitMQ.QueueDeclare(
		"retry.webhook.queue-3", // queue name
		true,                    // durable
		false,                   // auto delete
		false,                   // exclusive
		false,                   // no wait
		args,                    // arguments
	)
	if err != nil {
		panic(err)
	}

	_, err = channelRabbitMQ.QueueDeclare(
		"abaddon.webhook.queue", // queue name
		true,                    // durable
		false,                   // auto delete
		false,                   // exclusive
		false,                   // no wait
		args,                    // arguments
	)
	if err != nil {
		panic(err)
	}

	_ = channelRabbitMQ.QueueBind(
		"retry.webhook.queue-1",  // queue name
		"retry-1",                // routing key
		"retry.webhook.exchange", // exchange
		false,
		nil)

	_ = channelRabbitMQ.QueueBind(
		"retry.webhook.queue-2",  // queue name
		"retry-2",                // routing key
		"retry.webhook.exchange", // exchange
		false,
		nil)

	// Create a new Fiber instance.
	app := fiber.New()

	// Add middleware.
	app.Use(
		logger.New(), // add simple logger
	)

	// Add route for send message to Service 1.
	app.Get("/send", func(c *fiber.Ctx) error {
		// Create a message to publish.
		message := amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(c.Query("msg")),
		}

		queue := c.Query("queue")

		if queue == "1" {
			// Attempt to publish a message to the queue.
			if err := channelRabbitMQ.Publish(
				"retry.webhook.exchange", // exchange
				"retry-1",                // queue name
				false,                    // mandatory
				false,                    // immediate
				message,                  // message to publish
			); err != nil {
				return err
			}
		}
		if queue == "2" {
			// Attempt to publish a message to the queue.
			if err := channelRabbitMQ.Publish(
				"retry.webhook.exchange", // exchange
				"retry-2",                // queue name
				false,                    // mandatory
				false,                    // immediate
				message,                  // message to publish
			); err != nil {
				return err
			}
		}

		return nil
	})

	// Start Fiber API server.
	log.Fatal(app.Listen(":3000"))
}
