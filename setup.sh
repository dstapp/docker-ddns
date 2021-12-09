#!/bin/bash

[ -z "$SHARED_SECRET" ] && echo "SHARED_SECRET not set" && exit 1;
[ -z "$ZONE" ] && echo "ZONE not set" && exit 1;
[ -z "$RECORD_TTL" ] && echo "RECORD_TTL not set" && exit 1;

if ! grep 'zone "'$ZONE'"' /etc/bind/named.conf > /dev/null
then
	echo "creating zone...";
	cat >> /etc/bind/named.conf <<EOF
zone "$ZONE" {
	type master;
	file "$ZONE.zone";
	allow-query { any; };
EOF

if [ ! -z $TRANSFERIP ]
then
	cat >> /etc/bind/named.conf <<EOF
	allow-transfer { $TRANSFERIP; };
EOF
else 
	cat >> /etc/bind/named.conf <<EOF
	allow-transfer { none; };
EOF
fi
	cat >> /etc/bind/named.conf <<EOF
	allow-update { localhost; };
};
EOF
fi

if [ ! -f /var/cache/bind/$ZONE.zone ]
then
	echo "creating zone file..."
	cat > /var/cache/bind/$ZONE.zone <<EOF
\$ORIGIN .
\$TTL 86400	; 1 day
$ZONE		IN SOA	localhost. root.$ZONE. (
				74         ; serial
				3600       ; refresh (1 hour)
				900        ; retry (15 minutes)
				604800     ; expire (1 week)
				86400      ; minimum (1 day)
				)
EOF

if [ ! -z $NS ]
then
array=(${NS//,/ })
for var in ${array[@]}
do
cat >> /var/cache/bind/$ZONE.zone <<EOF
			NS	$var
EOF
done
else
cat >> /var/cache/bind/$ZONE.zone <<EOF
			NS	localhost
EOF
fi

cat >> /var/cache/bind/$ZONE.zone <<EOF
\$ORIGIN ${ZONE}.
\$TTL ${RECORD_TTL}
EOF
fi

# If /var/cache/bind is a volume, permissions are probably not ok
chown root:bind /var/cache/bind
chown bind:bind /var/cache/bind/*
chmod 770 /var/cache/bind
chmod 644 /var/cache/bind/*

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