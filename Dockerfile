FROM golang:1.15 AS builder

# Add your files into the container
ADD . /opt/build
WORKDIR /opt/build

# build the binary
RUN CGO_ENABLED=0 go build -v
FROM alpine:3.12
WORKDIR /

# COPY binary from previous stage to your desired location
COPY --from=builder /opt/build/consumer_prometheus .
ENV KAFKA_BROKERS="kafka01.bwnf.belwue.de:9093,kafka02.bwnf.belwue.de:9093,kafka03.bwnf.belwue.de:9093,kafka03.bwnf.belwue.de:9094,kafka05.bwnf.belwue.de:9093"
ENTRYPOINT /consumer --kafka.brokers $KAFKA_BROKERS --kafka.topic $KAFKA_TOPIC --kafka.consumer_group $KAFKA_CONSUMER_GROUP --kafka.user $KAFKA_USER --kafka.pass $KAFKA_PASS
