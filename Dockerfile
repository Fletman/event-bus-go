FROM golang:1.26-alpine AS builder
WORKDIR /build
COPY ./ /build/
RUN go build -o event-bus event-bus.go

FROM scratch
WORKDIR /app
COPY --from=builder /build/event-bus /app/event-bus
CMD [ "/app/event-bus" ]