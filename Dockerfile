FROM apline:3.21

COPY bin/dirbackup /usr/local/bin/dirbackup

RUN /usr/local/bin/dirbackup


