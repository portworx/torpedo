
version := 0.1.1

build: 	build-consul-bench \
		build-consul-agent	

push: 	push-consul-bench \
		push-consul-agent

build-consul-bench:
	docker build ./consul-bench --platform=linux/amd64 -t portworx/pds-loadtests:consul-bench-$(version)

push-consul-bench:
	docker push portworx/pds-loadtests:consul-bench-$(version)

build-consul-agent:
	docker build ./consul-agent --platform=linux/amd64 -t portworx/pds-loadtests:consul-agent-$(version)

push-consul-agent:
	docker push portworx/pds-loadtests:consul-agent-$(version)
