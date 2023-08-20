GOOS_LINUX = linux
GOOS_WINDOWS = windows
GOOS_MACOS = darwin

GOARCH_LINUX = amd64
GOARCH_WINDOWS = amd64
GOARCH_MACOS = arm64
GOARCH_MACOS_INTEL = amd64


install_linux:
	env GOOS=${GOOS_LINUX} GOARCH=${GOARCH_LINUX} ./builder.sh

install_macos:
	env GOOS=${GOOS_MACOS} GOARCH=${GOARCH_MACOS} ./builder.sh

install_macos_intel:
	env GOOS=${GOOS_MACOS} GOARCH=${GOARCH_MACOS_INTEL} ./builder.sh

uninstall:
	./uninstall.sh

unit_test_api:
	cd api/ && \
		chmod u+x test-api.sh && \
		./test-api.sh

mock_civo:
	rm -rf ~/.ksctl && \
	mkdir -p ~/.ksctl/config/civo/managed && \
	mkdir -p ~/.ksctl/config/civo/ha

	cd test/ && go test -bench=BenchmarkCivoTestingHA -benchtime=1x -cover

test: unit_test_api mock_civo
	echo "Done"

