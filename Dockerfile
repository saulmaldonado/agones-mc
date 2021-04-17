FROM golang:alpine3.13 AS build
WORKDIR /agones-mc-monitor/
COPY . .
RUN CGO_ENABLED=0 go build -o ./agones-mc-monitor ./cmd/main.go

FROM scratch
WORKDIR /agones-mc-monitor/
COPY --from=build /agones-mc-monitor/agones-mc-monitor .
ENTRYPOINT [ "./agones-mc-monitor" ]
