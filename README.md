# Setting up OS

We are using confluentic official go kafka package, so this package uses kafka *librdkafka* (C client) internally and because of that you will need to install some deps on your OS

For ubuntu users

`sudo apt-get install build-essential pkg-config git`

# Running Confluent On Docker Compose
Community Edition : `make run-community` 
Commercial Edition : `make run-commercial` 

# Running Confluent On Kubernetes (K8s)

# Examples (on folder proof-of-concept)
- ex. confluent-kafka-go-3
`go run main.go`

Note : 
*kafdrop* will run on port `29000` and *kafka server* will run on `9092`
