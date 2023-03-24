FROM alpine:3.17.2 AS alpine0

RUN mkdir /whisper && \
  wget -q https://github.com/masterful/whisper.cpp/tarball/master -O - | \
  tar -xz -C /whisper --strip-components 1


WORKDIR /whisper

RUN apk add --quiet g++ make bash wget sdl2-dev alsa-utils

# Workaround build
RUN make main stream command talk bench

# Once workaround is no longer necessary
# RUN make main stream command talk bench

RUN bash ./models/download-ggml-model.sh base.en

FROM alpine:3.17.2

COPY --from=alpine0 /whisper/main /usr/local/bin/whisper
COPY --from=alpine0 /whisper/stream /usr/local/bin/stream
COPY --from=alpine0 /whisper/talk /usr/local/bin/talk
COPY --from=alpine0 /whisper/command /usr/local/bin/wcommand
RUN mkdir /root/models
COPY --from=alpine0 /whisper/models/ggml-base.en.bin /root/models/ggml-base.en.bin

RUN apk add --quiet sdl2-dev alsa-utils

WORKDIR /root

# To use stream in a container (linux hosts only):
# docker build . -t whisper
# docker run -it --rm --device /dev/snd:/dev/snd whisper stream

