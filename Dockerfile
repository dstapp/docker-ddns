FROM debian:stretch
MAINTAINER David Prandzioch <hello+ddns@davd.eu>

RUN DEBIAN_FRONTEND=noninteractive apt-get update && \
	apt-get install -q -y bind9 dnsutils golang git-core && \
	apt-get clean

RUN chmod 770 /var/cache/bind

COPY setup.sh /root/setup.sh
RUN chmod +x /root/setup.sh

ENV GOPATH=/root/go
RUN mkdir -p /root/go/src
COPY rest-api /root/go/src/dyndns
RUN cd /root/go/src/dyndns && go get

COPY named.conf.options /etc/bind/named.conf.options

EXPOSE 53 8080
CMD ["sh", "-c", "/root/setup.sh ; service bind9 start ; /root/go/bin/dyndns"]
