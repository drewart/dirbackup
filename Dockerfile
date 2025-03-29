FROM alpine:3.21

COPY bin/* /usr/local/bin/

CMD [ "sh", "-c", "/usr/local/bin/dirbackup-service" ]


