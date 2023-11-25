GOOS_LINUX = linux
GOOS_WINDOWS = windows
GOOS_MACOS = darwin

GOARCH_LINUX = amd64
GOARCH_WINDOWS = amd64
GOARCH_MACOS = arm64
GOARCH_MACOS_INTEL = amd64

CURR_TIME = $(shell date +%s)

docker_build_httpserver:
	docker build --file build/httpserver_slim/Dockerfile --tag docker.io/kubesimplify/ksctl:slim-v1 .

docker_push_registry_httpserver:
	docker push docker.io/kubesimplify/ksctl:slim-v1

createFolders:
	mkdir -p ${HOME}/.ksctl/cred
	mkdir -p ${HOME}/.ksctl/config/civo/ha
	mkdir -p ${HOME}/.ksctl/config/civo/managed
	mkdir -p ${HOME}/.ksctl/config/azure/ha
	mkdir -p ${HOME}/.ksctl/config/azure/managed
	mkdir -p ${HOME}/.ksctl/config/aws/managed
	mkdir -p ${HOME}/.ksctl/config/aws/ha
	mkdir -p ${HOME}/.ksctl/config/local/managed
	@echo "Configuration folders setup created done"

deleteFolders:
	rm -rf ${HOME}/.ksctl
	@echo "Configuration folders setup deleted done"

install_linux:
	@echo "Started to Install ksctl"
	cd scripts && \
		env GOOS=${GOOS_LINUX} GOARCH=${GOARCH_LINUX} ./builder.sh

install_macos:
	@echo "Started to Install ksctl"
	cd scripts && \
		env GOOS=${GOOS_MACOS} GOARCH=${GOARCH_MACOS} ./builder.sh

install_macos_intel:
	@echo "Started to Install ksctl"
	cd scripts && \
		env GOOS=${GOOS_MACOS} GOARCH=${GOARCH_MACOS_INTEL} ./builder.sh

uninstall:
	@echo "Started to Uninstall ksctl"
	cd scripts && \
		./uninstall.sh

unit_test_api:
	@echo "Unit Tests"
	cd scripts/ && \
		chmod u+x test-api.sh && \
		./test-api.sh

mock_test:
	@echo "Mock Test (integration)"
	cd test/ && \
		go test -bench=. -benchtime=1x -cover -v

mock_civo_ha:
	cd test/ && \
 		go test -bench=BenchmarkCivoTestingHA -benchtime=1x -cover -v

mock_civo_managed:
	cd test/ && \
 		go test -bench=BenchmarkCivoTestingManaged -benchtime=1x -cover -v

mock_azure_managed:
	cd test/ && \
 		go test -bench=BenchmarkAzureTestingManaged -benchtime=1x -cover -v

mock_azure_ha:
	cd test/ && \
 		go test -bench=BenchmarkAzureTestingHA -benchtime=1x -cover -v

mock_aws_ha:
	cd test/ && \
		go test -bench=BemchmarkAwsTestingHA -benchtime=1x -cover -v

mock_local_managed:
	cd test/ && \
 		go test -bench=BenchmarkLocalTestingManaged -benchtime=1x -cover -v

test: unit_test_api mock_test
	@echo "Done All tests"

