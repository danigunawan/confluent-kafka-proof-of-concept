# COMMUNITY
https://github.com/confluentinc/cp-all-in-one
image: confluentinc/cp-kafka:latest # Community Edition

# COMMERCIAL
image: confluentinc/cp-server:7.0.0 # Commercial Edition

# EXAMPLES 
https://github.com/confluentinc/examples

*kafdrop* will run on port `29000` and *kafka server* will run on `9092`

# go-kafka-producer (example client : confluent-kafka-go-2)

`go run main.go -topic topic_name -message "some cool message"`

# Setting up OS

We are using confluentic official go kafka package, so this package uses kafka *librdkafka* (C client) internally and because of that you will need to install some deps on your OS

For ubuntu users

`sudo apt-get install build-essential pkg-config git`
