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

mock_test:
	cd test/ && go test -bench=. -benchtime=1x -cover -v


mock_civo_ha:
	cd test/ && go test -bench=BenchmarkCivoTestingHA -benchtime=1x -cover -v


mock_civo_managed:
	cd test/ && go test -bench=BenchmarkCivoTestingManaged -benchtime=1x -cover -v


mock_azure_managed:
	cd test/ && go test -bench=BenchmarkAzureTestingManaged -benchtime=1x -cover -v

mock_azure_ha:
	cd test/ && go test -bench=BenchmarkAzureTestingHA -benchtime=1x -cover -v

test: unit_test_api mock_test
	echo "Done"

