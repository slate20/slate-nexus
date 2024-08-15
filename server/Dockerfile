FROM checkmk/check-mk-raw:latest

RUN apt-get update && apt-get install -y socat

CMD /bin/bash -c "socat TCP-LISTEN:6557,reuseaddr,fork UNIX-CLIENT:/omd/sites/main/tmp/run/live & exec /docker-entrypoint.sh"