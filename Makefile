image:
	docker build -t davd/dyndns-server .

console:
	docker run -it -p 8080:8080 -p 53:53 -p 53:53/udp --rm davd/dyndns-server bash

server_test:
	docker run -it -p 8080:8080 -p 53:53 -p 53:53/udp --env-file envfile --rm davd/dyndns-server

api_test:
	curl "http://localhost:8080/update?secret=changeme&domain=foo&addr=1.2.3.4"
	dig @localhost foo.example.org

api_test_recursion:
	dig @localhost google.com

deploy: image
	docker run -it -d -p 8080:8080 -p 53:53 -p 53:53/udp --env-file envfile --name=dyndns davd/dyndns-server