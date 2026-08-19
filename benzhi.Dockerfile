FROM golang:1.21

ENV GOTOOLCHAIN=local
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY go.mod ./
COPY . .

RUN go build -o /app/bin .

CMD ["bash"]
