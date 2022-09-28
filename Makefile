
docker_builder:
	docker build -t kubesimpctl -f build/Dockerfile src/cli/

docker_clean:
	docker rmi -f kubesimpctl
