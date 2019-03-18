[1.3.0]
* Add basic CI integration
* Add usage example for docker-compose
* Use request IP address when no `addr` is provided for better compatibility with DD-WRT (by vdweegen)
* Allow IPv4 and IPv6 addresses to co-exist for the same domain

[1.2.1]
* Fix permissions of /var/cache/bind on container startup
* Create zone options if not done, fixing support for persistent volumes

[1.2.0]
* Allow usage of multiple domains

[1.1.0]
* Update Debian Jessie to Debian Stretch
* Multistage Dockerfile resulting in smaller production image
* Code refactoring
* Extended response
* Basic unit test coverage
* Documentation on running from DockerHub

[1.0.0]
* Initial release
