FROM golang:1.19-alpine as alpine0

RUN mkdir /whisper && \
  wget -q https://github.com/masterful/whisper.cpp/tarball/master -O - | \
  tar -xz -C /whisper --strip-components 1

WORKDIR /whisper/
RUN apk add --quiet g++ make bash wget sdl2-dev alsa-utils
RUN make main talk CFLAGS=-D_POSIX_C_SOURCE=199309L
RUN bash ./models/download-ggml-model.sh base.en

# Set the Current Working Directory inside the container
WORKDIR /opt/

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .

# Download all the dependencies
RUN go mod download

# Install the package
RUN go build -o chatgptgo


FROM alpine:3.17.2

COPY --from=alpine0 /opt/chatgptgo /usr/local/bin/chatgptgo
COPY --from=alpine0 /whisper/main /usr/local/bin/whisper
#COPY --from=alpine0 /whisper/stream /usr/local/bin/stream
COPY --from=alpine0 /whisper/talk /usr/local/bin/talk
#COPY --from=alpine0 /whisper/command /usr/local/bin/wcommand
RUN mkdir /root/models
COPY --from=alpine0 /whisper/models/ggml-base.en.bin /root/models/ggml-base.en.bin

RUN apk add --quiet sdl2-dev alsa-utils

CMD ["chatgptgo"]
