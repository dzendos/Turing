FROM golang:1.19

RUN apt install -y git

WORKDIR /usr/src/app

RUN git clone https://github.com/dzendos/Turing

WORKDIR /usr/src/app/Turing/config
COPY config/config.json .

WORKDIR /usr/src/app/Turing
COPY begin.sh .


# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify
RUN go get gopkg.in/tucnak/telebot.v2
RUN go build -v -o /usr/src/app/Turing ./...

CMD sh /usr/src/app/Turing/begin.sh