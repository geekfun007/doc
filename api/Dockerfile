FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY go.work go.work
COPY kitex_gen/ kitex_gen/
COPY pkg/ pkg/
COPY api/ api/

RUN cd api && go build -o /build/api-server .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /build/api-server .

EXPOSE 8080

CMD ["./api-server"]
