FROM debian:9
MAINTAINER jemoster@gmail.com

RUN echo 'deb http://deb.debian.org/debian buster main' >> /etc/apt/sources.list
RUN echo 'deb http://deb.debian.org/debian buster-updates main' >> /etc/apt/sources.list
RUN cat /etc/apt/sources.list
RUN apt update

RUN apt install -y python3.6
