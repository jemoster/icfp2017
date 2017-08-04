# Use Debian 9 as the base since this is what the competition is using.
FROM debian:9
MAINTAINER jemoster@gmail.com

# Hack in the Debian 10 "buster" repos since these contain more up to date
# versions of the packages to align with the competition spec.
RUN echo 'deb http://deb.debian.org/debian buster main' >> /etc/apt/sources.list
RUN echo 'deb http://deb.debian.org/debian buster-updates main' >> /etc/apt/sources.list
RUN cat /etc/apt/sources.list
RUN apt update

RUN apt install -y python3.6
