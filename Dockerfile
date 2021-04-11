FROM golang:alpine3.13

WORKDIR /agones-mc-monitor/

COPY . .

RUN CGO_ENABLED=0 go build -o ./agones-mc-monitor ./main.go

ENTRYPOINT [ "./agones-mc-monitor" ]
