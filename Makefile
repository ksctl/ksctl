GOOS_LINUX = linux
GOOS_WINDOWS = windows
GOOS_MACOS = darwin

GOARCH_LINUX = amd64
GOARCH_WINDOWS = amd64
GOARCH_MACOS = arm64
GOARCH_MACOS_INTEL = amd64

unit_test:
	cd test/api && \
	./civo_test.sh && \
	./local_test.sh

install_linux:
	env GOOS=${GOOS_LINUX} GOARCH=${GOARCH_LINUX} ./install.sh

install_macos:
	env GOOS=${GOOS_MACOS} GOARCH=${GOARCH_MACOS} ./install.sh

install_macos_intel:
	env GOOS=${GOOS_MACOS} GOARCH=${GOARCH_MACOS_INTEL} ./install.sh

uninstall:
	./uninstall.sh
