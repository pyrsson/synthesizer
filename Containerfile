FROM golang:1.22-alpine as builder

COPY . /src
WORKDIR /src
RUN go build -o /synthesizer

FROM alpine:3.19
COPY --from=builder /synthesizer /bin/synthesizer
CMD ["/bin/synthesizer"]
