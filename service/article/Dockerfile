FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY go.work go.work
COPY kitex_gen/ kitex_gen/
COPY pkg/ pkg/
COPY service/article/ service/article/

RUN cd service/article && go build -o /build/article-server .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /build/article-server .

EXPOSE 8882

CMD ["./article-server"]
