# How to run the core for Testing or Trying out the latest core functionalties

## Prerequisites
- make sure you had ran both the unit test cases and mock test cases. Refer
```bash
make help
```

- if the controller logic is changing especially inside the `ksctl-components/operators/**` then make sure you had ran the following commands
```bash
make generate CONTROLLER=application
make manifests CONTROLLER=application
```

- If there are any changes in the `ksctl-components/operators/**` make sure you had ran the `make build-installer` command to generate the latest manifest `ksctl-components/operators/application` image.

---

> [!NOTE]
> Working Directory is `test/e2e`

> [!IMPORTANT]
> when specifying the env variable `E2E_FLAGS` make sure you are not using the `;` in the `core_component_overridings` value.

### If Working on Main branch

```bash
go build -v .
./e2e -op create -file local/create.json
./e2e -op switch -file local/switch.json
./e2e -op delete -file local/delete.json
```

### If working on a PR

you will get the instrction from the PR having `pr/lgtm` label added. ask the maintainers to add the label if it is not there.
Condition that the PR is done from the testing and all

```bash
make build-installer CONTROLLER="application" CUSTOM_LOCATION_GEN_CONTROLLER_MANIFEST="/tmp/ksctl-manifest.yaml" IMG_SUFFIX="<GET FROM GH ACTION PR COMMENT>" IMG_TAG_VERSION="<GET FROM GH ACTION PR COMMENT>"

cd test/e2e
go build -v -ldflags="-X 'github.com/ksctl/ksctl/commons.OCIVersion=<GET FROM GH ACTION PR COMMENT>' -X 'github.com/ksctl/ksctl/commons.OCIImgSuffix=<GET FROM GH ACTION PR COMMENT>'" .

export E2E_FLAGS="debug;core_component_overridings=application=file:::/tmp/ksctl-manifest.yaml" ## you can eliminate the debug

./e2e -op create -file local/create.json
./e2e -op switch -file local/switch.json
./e2e -op delete -file local/delete.json
```
