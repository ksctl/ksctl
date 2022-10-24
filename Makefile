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
	docker build -t ksctl -f build/Dockerfile src/cli/

docker_run:
	docker run --rm -it ksctl bash

docker_clean:
	docker rmi -f ksctl

build_exec:
	cd src/cli && go build -v -o ksctl .

rm_exec:
	rm -vf src/cli/ksctl
