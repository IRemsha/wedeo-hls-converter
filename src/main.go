package main

// Import Go and NATS packages
import (
    "fmt"
    "runtime"

    "github.com/nats-io/stan.go"
)

func main() {
    var stanConnection stan.Conn

    subscribe := func() {
        fmt.Printf("Subscribing to subject 'bucketevents'\n")
        stanConnection.Subscribe("bucketevents", func(m *stan.Msg) {

            // Handle the message
            fmt.Printf("Received a message: %s\n", string(m.Data))
        })
    }


    stanConnection, _ = stan.Connect("test-cluster", "test-client", stan.NatsURL("nats://yourusername:yoursecret@0.0.0.0:4222"), stan.SetConnectionLostHandler(func(c stan.Conn, _ error) {
        go func() {
            for {
                // Reconnect if the connection is lost.
                if stanConnection == nil || stanConnection.NatsConn() == nil ||  !stanConnection.NatsConn().IsConnected() {
                    stanConnection, _ = stan.Connect("test-cluster", "test-client", stan.NatsURL("nats://yourusername:yoursecret@0.0.0.0:4222"), stan.SetConnectionLostHandler(func(c stan.Conn, _ error) {
                        if c.NatsConn() != nil {
                            c.NatsConn().Close()
                        }
                        _ = c.Close()
                    }))
                    if stanConnection != nil {
                        subscribe()
                    }

                }
            }

        }()
    }))

    // Subscribe to subject
    subscribe()

    // Keep the connection alive
    runtime.Goexit()
}
