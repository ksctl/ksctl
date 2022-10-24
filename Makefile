GOOS_LINUX = linux
GOOS_WINDOWS = windows
GOOS_MACOS = darwin

GOARCH_LINUX = amd64
GOARCH_WINDOWS = amd64
GOARCH_MACOS = arm64


install_linux:
	env GOOS=${GOOS_LINUX} GOARCH=${GOARCH_LINUX} ./install-linux.sh

uninstall_linux:
	./uninstall-linux.sh

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
