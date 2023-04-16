FROM gcr.io/distroless/static

ADD bot /usr/local/bin/bot
ADD wrapper /usr/local/bin/wrapper

ENTRYPOINT ["wrapper"]
