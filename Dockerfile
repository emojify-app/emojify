FROM alpine

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

RUN mkdir /service

COPY ./emojify-server /service/
COPY ./images /service/images/

WORKDIR /service

ENTRYPOINT /service/emojify-server
