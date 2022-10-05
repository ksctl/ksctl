
docker_builder:
	docker build -t kubesimpctl -f build/Dockerfile src/cli/

docker_run:
	docker run --rm -it kubesimpctl bash

docker_clean:
	docker rmi -f kubesimpctl

build_exec:
	cd src/cli && go build -v -o kubesimpctl .

rm_exec:
	rm -vf src/cli/kubesimpctl