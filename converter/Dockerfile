# Start with the Linuxserver ffmpeg image
FROM linuxserver/ffmpeg:latest

RUN apt update && apt upgrade -y

RUN mkdir /Conversions

COPY ./go-converter /app

ENTRYPOINT [ "/app/go-converter" ]
