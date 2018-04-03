author = samuelmyers
version = 001
wp-base = http://objects.liquidweb.services/triv/mwch-base.tar.gz

all: extract mysql wordpress ping

extract:
	wget $(wp-base)
	tar -xvzf mwch-base.tar.gz

mysql:
	docker build -t $(author)/mysql-dump:$(version) -f mysql-dump.dockerfile .
	docker push $(author)/mysql-dump:$(version)

wordpress:
	docker build -t $(author)/wordpress-extract:$(version) -f wordpress-extract.dockerfile .
	docker push $(author)/wordpress-extract:$(version)

ping:
	docker build -t $(author)/ping-test:$(version) -f ping-test.dockerfile .
	docker push $(author)/ping-test:$(version)
