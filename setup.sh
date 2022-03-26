#!/bin/bash
NAMED_HOST=${NAMED_HOST:-'localhost'}
ZONES=$(echo $ZONE | tr ',' '\n')
RECORD_TTL=${RECORD_TTL:-300}

# Backward compatibility for a single zone
if [ $(echo "$ZONES" | wc -l) -eq 1 ]; then
    # Allow update without fqdn
    ZONE=$(echo "$ZONES" | head -1)
else
    # Allow multiple zones and disable updates without fqdn
    ZONE=""
fi

# replaces a config value
function bind-conf-set {
    local KEY=${1:?'No key set'}
    local SECRED=${2:-$(openssl rand 32 | base64)}
    sed -E 's@('"${KEY}"')(\W+)"(.*)"@\1 "'${SECRED}'"@g' /dev/stdin
}

function bind-zone-add {
local ZONE=${1:?'No zone set'}
if ! grep 'zone "'$ZONE'"' /etc/bind/named.conf > /dev/null
then
	echo "creating zone for $ZONE...";
	cat >> /etc/bind/named.conf <<EOF
include "/etc/bind/ddns/$ZONE.key";
zone "$ZONE" {
	type master;
	file "$ZONE.zone";
	allow-query { any; };
	allow-transfer { none; };
	update-policy { grant ddns-key.$ZONE zonesub ANY; };
};
EOF
fi

if [ ! -f /var/cache/bind/$ZONE.zone ]
then
	echo "creating zone file for $ZONE..."
	cat > /var/cache/bind/$ZONE.zone <<EOF
\$ORIGIN ${ZONE}.
\$TTL 86400	; 1 day
@		IN SOA	ns postmaster (
				74         ; serial
				3600       ; refresh (1 hour)
				900        ; retry (15 minutes)
				604800     ; expire (1 week)
				86400      ; minimum (1 day)
				)
			NS	ns
ns			A	127.0.0.1
EOF
fi
}

install -d -o bind -g bind /etc/bind/ddns

for Z in $ZONES; do
    ddns-confgen -q -z $Z | bind-conf-set secret "${SHARED_SECRET}" | tee /etc/bind/ddns/$Z.key
    bind-zone-add $Z
done

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
    "Server": "${NAMED_HOST}",
    "Zone": "${ZONE}.",
    "NsupdateBinary": "/usr/bin/nsupdate",
	"RecordTTL": ${RECORD_TTL}
}
EOF
fi