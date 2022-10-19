# syntax=docker/dockerfile:1
FROM golang:1.18 as builder

# build k6 0.36.0 with prometheus support
ENV K6_VERSION=v0.36.0
RUN go install go.k6.io/xk6/cmd/xk6@v0.6.1 && xk6 build --with github.com/grafana/xk6-output-prometheus-remote@latest
RUN  xk6 build --output "/tmp/k6" --with github.com/grafana/xk6-browser

FROM debian:bullseye

ARG CHROMIUM_VERSION=106.0.5249.61-1~deb11u1


RUN apt install -y chromium

WORKDIR /build

ENV CGO_ENABLED=0 
ENV GOOS=linux
COPY . .


FROM loadimpact/k6:0.36.0
WORKDIR /home/k6
COPY --from=builder /go/k6 /usr/bin/k6
COPY --from=builder /build/runner /bin/runner
ENV XK6_HEADLESS=true
ENTRYPOINT ["/bin/runner"]
