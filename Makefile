image:
	docker build -t davd/docker-ddns:latest .

console:
	docker run -it -p 8080:8080 -p 53:53 -p 53:53/udp --rm davd/docker-ddns:latest bash

devconsole:
	docker run -it --rm -v ${PWD}/rest-api:/usr/src/app -w /usr/src/app golang:1.8.5 bash

server_test:
	docker run -it -p 8080:8080 -p 53:53 -p 53:53/udp --env-file envfile --rm davd/docker-ddns:latest

unit_tests:
	docker run -it --rm -v ${PWD}/rest-api:/go/src/dyndns -w /go/src/dyndns golang:1.8.5 /bin/bash -c "go get && go test -v"

api_test:
	curl "http://docker.local:8080/update?secret=changeme&domain=foo&addr=1.2.3.4"
	dig @docker.local foo.example.org

api_test_invalid_params:
	curl "http://docker.local:8080/update?secret=changeme&addr=1.2.3.4"
	dig @docker.local foo.example.org

api_test_recursion:
	dig @docker.local google.com

deploy: image
	docker run -it -d -p 8080:8080 -p 53:53 -p 53:53/udp --env-file envfile --name=dyndns davd/docker-ddns:latest
