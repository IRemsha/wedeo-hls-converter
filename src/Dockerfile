FROM golang:1.17

RUN mkdir -p /src/

RUN cd /src/ && \
    apt-get -y update && \
    apt-get -y upgrade && \
    apt-get install -y ffmpeg

COPY . /src/
WORKDIR /src/

CMD [ "go", "run", "main.go" ]
