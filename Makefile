image:
	docker build -t davd/docker-ddns:latest .

console:
	docker run -it -p 8080:8080 -p 53:53 -p 53:53/udp --rm davd/docker-ddns:latest bash

devconsole:
	docker run -it --rm -v ${PWD}/rest-api:/usr/src/app -w /usr/src/app golang:1.14 bash

server_test:
	docker run -it -p 8080:8080 -p 53:53 -p 53:53/udp --env-file envfile --rm davd/docker-ddns:latest

unit_tests:
	docker run -it --rm -v ${PWD}/rest-api:/go/src/dyndns -w /go/src/dyndns golang:1.14 /bin/bash -c "go get && go test -v"

api_test:
	curl "http://localhost:8080/update?secret=changeme&domain=foo&addr=1.2.3.4"
	dig @localhost foo.example.org

api_test_46:
	curl "http://localhost:8080/update?secret=changeme&domain=foo&addr=1.2.3.4"
	curl "http://localhost:8080/update?secret=changeme&domain=foo&addr=2001:0db8:85a3:0000:0000:8a2e:0370:7334"
	dig @localhost foo.example.org
	dig @localhost AAAA foo.example.org

api_test_multiple_domains:
	curl "http://localhost:8080/update?secret=changeme&domain=foo,bar,baz&addr=1.2.3.4"
	dig @localhost foo.example.org
	dig @localhost bar.example.org
	dig @localhost baz.example.org

api_test_invalid_params:
	curl "http://localhost:8080/update?secret=changeme&addr=1.2.3.4"
	dig @localhost foo.example.org

api_test_recursion:
	dig @localhost google.com

deploy: image
	docker run -it -d -p 8080:8080 -p 53:53 -p 53:53/udp --env-file envfile --name=dyndns davd/docker-ddns:latest
