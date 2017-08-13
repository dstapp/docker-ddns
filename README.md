# Dynamic DNS with Docker, Go and Bind9

This package allows you to set up a server for dynamic DNS using docker with a
few simple commands. You don't have to worry about nameserver setup, REST API
and all that stuff.

## Installation

You can either take the image from DockerHub or build it on your own.

### Using DockerHub

Just customize this to your needs and run:

```
docker run -it -d \
    -p 8080:8080 \
    -p 53:53 \
    -p 53:53/udp \
    -e SHARED_SECRET=changeme \
    -e ZONE=example.org \
    -e RECORD_TTL=3600 \
    --name=dyndns \
    davd/docker-ddns:latest
```

### Build from source / GitHub

```
git clone https://github.com/dprandzioch/docker-ddns
cd docker-ddns
$EDITOR envfile
make deploy
```

Make sure to change all environment variables in `envfile` to match your needs. Some more information can be found here: https://www.davd.eu/build-your-own-dynamic-dns-in-5-minutes/

## Exposed ports

Afterwards you have a running docker container that exposes three ports:

* 53/TCP    -> DNS
* 53/UDP    -> DNS
* 8080/TCP  -> Management REST API


## Using the API

That package features a simple REST API written in Go, that provides a simple
interface, that almost any router that supports Custom DDNS providers can
attach to (e.g. Fritz!Box). It is highly recommended to put a reverse proxy
before the API.

It provides one single GET request, that is used as follows:

http://myhost.mydomain.tld:8080/update?secret=changeme&domain=foo&addr=1.2.3.4

### Fields

* `secret`: The shared secret set in `envfile`
* `domain`: The subdomain to your configured domain, in this example it would
result in `foo.example.org`
* `addr`: IPv4 or IPv6 address of the name record

## Accessing the REST API log

Just run

```
docker logs -f dyndns
```

## DNS setup

To provide a little help... To your "real" domain, like `domain.tld`, you
should add a subdomain that is delegated to this DDNS server like this:

```
dyndns                   IN NS      ns
ns                       IN A       <put ipv4 of dns server here>
ns                       IN AAAA    <optional, put ipv6 of dns server here>
```

Your management API should then also be accessible through

```
http://ns.domain.tld:8080/update?...
```

If you provide `foo` as a domain when using the REST API, the resulting domain
will then be `foo.dyndns.domain.tld`.
