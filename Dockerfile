FROM alpine:3.21

COPY bin/dirbackup /usr/local/bin/dirbackup

COPY bin/dirsynctime /usr/local/bin/dirsynctime

CMD /usr/local/bin/dirbackup


