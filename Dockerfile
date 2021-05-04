FROM golang:1.16.3 as build

WORKDIR /agones-mc/

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o ./build/agones-mc main.go

FROM scratch
WORKDIR /agones-mc/
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /agones-mc/build/agones-mc .
ENTRYPOINT [ "./agones-mc" ]
