FROM golang:1.21-alpine
ENV GOTOOLCHAIN=local
ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /app
COPY . .
RUN go build -o /app/bin .
CMD ["/app/bin", "--help"]
