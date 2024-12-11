FROM golang:1.23 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY agent.go agent.go
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/agent 

FROM quay.io/fedora/fedora-bootc:40
LABEL bootc-agent-enabled="true"
COPY --from=build /app/bin/agent /usr/bin
ADD rabbitmq-listener.service /usr/lib/systemd/system/
COPY amqp-config.yaml /etc
RUN ln -s rabbitmq-listener.service /usr/lib/systemd/system/multi-user.target.wants