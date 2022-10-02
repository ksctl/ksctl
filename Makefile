
docker_builder:
	docker build -t kubesimpctl -f build/Dockerfile src/cli/

docker_run:
	docker run --rm -it kubesimpctl bash

docker_clean:
	docker rmi -f kubesimpctl
