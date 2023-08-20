package civo_test

import (
	"github.com/kubesimplify/ksctl/api/provider/civo"
)

func ProvideMockCivoClient() civo.CivoGo {
	return &civo.CivoGoMockClient{}
}
