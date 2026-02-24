FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY go.work go.work
COPY kitex_gen/ kitex_gen/
COPY pkg/ pkg/
COPY service/user/ service/user/

RUN cd service/user && go build -o /build/user-server .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /build/user-server .

EXPOSE 8881

CMD ["./user-server"]
