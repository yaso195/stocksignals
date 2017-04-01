FROM alpine:latest

MAINTAINER Edward Muller <edward@heroku.com>

WORKDIR "/opt"

ADD .docker_build/stocksignals /opt/bin/stocksignals

CMD ["/opt/bin/stocksignals"]

