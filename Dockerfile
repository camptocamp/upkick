FROM golang:1.15 as builder
WORKDIR /opt/build
COPY . .
RUN make build

FROM scratch
WORKDIR /
COPY --from=builder /opt/build/upkick /
ENTRYPOINT ["/upkick"]
