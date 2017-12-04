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
	file "$ZONE.zone.signed";
	allow-query { any; };
	allow-transfer { none; };
	allow-update { localhost; };
};
EOF
	
	echo "creating zone file..."
	if [ 'z "$NS" ]
	then
		IFS="," read -r -a elements <<< "$NS"
		for element in ${elements[@]}
		do
			SHORT+="${element%%.*}. "
			LONG+="$element. "
		done
	else
		SHORT="${NS%%.*}."
		LONG+="$NS."
	fi
	cat > /var/cache/bind/$ZONE.zone <<EOF
\$ORIGIN .
\$TTL 86400	; 1 day
$ZONE		IN SOA	$SHORT $LONG (
				74         ; serial
				3600       ; refresh (1 hour)
				900        ; retry (15 minutes)
				604800     ; expire (1 week)
				86400      ; minimum (1 day)
				)
			NS	$SHORT
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

chown -R bind:bind /var/cache/bind

# DNSSEC configuration
if [ ! -f /var/cache/bind/$ZONE.zone.signed ]
then
    echo "Signing zone..."
    cd /var/cache/bind
    dnssec-keygen -a NSEC3RSASHA1 -b 2048 -n ZONE $ZONE
    dnssec-keygen -f KSK -a NSEC3RSASHA1 -b 4096 -n ZONE $ZONE
    for key in `ls K${ZONE}*.key`
    do
        echo "\$INCLUDE $key">> $ZONE.zone
    done

    dnssec-signzone -A -3 $(head -c 1000 /dev/urandom | sha1sum | cut -b 1-16) -N INCREMENT -o $ZONE -t $ZONE.zone
fi

# Increase safety to prevents hacks with raindow tables
if [ ! -f /usr/sbin/zonesigner.sh ]
then
    echo "Creating /usr/sbin/zonesigner.sh..."
    cat > /usr/sbin/zonesigner.sh <<EOF
#!/bin/sh

PDIR=\`pwd\`
ZONEDIR="/var/cache/bind" #location of your zone files
ZONE=\$1
ZONEFILE=\$2
DNSSERVICE="bind9" #On CentOS/Fedora replace this with "named"
cd \$ZONEDIR
SERIAL=\`/usr/sbin/named-checkzone \$ZONE \$ZONEFILE | egrep -ho '[0-9]{10}'\`
sed -i 's/'\$SERIAL'/'\$((\$SERIAL+1))'/' \$ZONEFILE
/usr/sbin/dnssec-signzone -A -3 \$(head -c 1000 /dev/urandom | sha1sum | cut -b 1-16) -N increment -o \$1 -t \$2
service \$DNSSERVICE reload
cd \$PDIR
EOF

    chmod +x /usr/sbin/zonesigner.sh
fi

if [ ! -f /var/spool/cron/crontabs/root ]
then
    echo "Implements crontab..."
    cat > /var/spool/cron/crontabs/root <<EOF
# Edit this file to introduce tasks to be run by cron.
#
# Each task to run has to be defined through a single line
# indicating with different fields when the task will be run
# and what command to run for the task
#
# To define the time you can provide concrete values for
# minute (m), hour (h), day of month (dom), month (mon),
# and day of week (dow) or use '*' in these fields (for 'any').#
# Notice that tasks will be started based on the cron's system
# daemon's notion of time and timezones.
#
# Output of the crontab jobs (including errors) is sent through
# email to the user the crontab file belongs to (unless redirected).
#
# For example, you can run a backup of all your user accounts
# at 5 a.m every week with:
# 0 5 * * 1 tar -zcf /var/backups/home.tgz /home/
#
# For more information see the manual pages of crontab(5) and cron(8)
#
# m h  dom mon dow   command
0 0 */3 0 0 /usr/sbin/zonesigner.sh $ZONE $ZONE.zone
EOF
fi

echo "Service Bind9 restart..."
service bind9 restart
