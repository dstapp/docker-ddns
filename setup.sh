#!/bin/bash

[ -z "$SHARED_SECRET" ] && echo "SHARED_SECRET not set" && exit 1;
[ -z "$ZONE" ] && echo "ZONE not set" && exit 1;
[ -z "$RECORD_TTL" ] && echo "RECORD_TTL not set" && exit 1;

if [ ! -f /var/cache/bind/$ZONE.zone ]
then
	echo "creating zone...";
	cat >> /etc/bind/named.conf <<EOF
zone "$ZONE" {
	type master;
	file "$ZONE.zone";
	allow-query { any; };
	allow-transfer { none; };
	allow-update { localhost; };
};
EOF
	
	echo "creating zone file..."
	cat > /var/cache/bind/$ZONE.zone <<EOF
\$ORIGIN .
\$TTL 86400	; 1 day
$ZONE		IN SOA	localhost. root.localhost. (
				74         ; serial
				3600       ; refresh (1 hour)
				900        ; retry (15 minutes)
				604800     ; expire (1 week)
				86400      ; minimum (1 day)
				)
			NS	localhost.
\$ORIGIN ${ZONE}.
\$TTL ${RECORD_TTL}
EOF
fi

if [ ! -f /etc/dyndns.json ]
then
	echo "creating REST api config..."
	cat > /etc/dyndns.json <<EOF
{
    "SharedSecret": "${SHARED_SECRET}",
    "Server": "localhost",
    "Zone": "${ZONE}.",
    "Domain": "${ZONE}",
    "NsupdateBinary": "/usr/bin/nsupdate",
	"RecordTTL": ${RECORD_TTL}
}
EOF
fi