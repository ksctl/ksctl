package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/kubesimplify/ksctl/pkg/utils/consts"
	"github.com/spf13/cobra"
)

const (
	ksctl_feature_auto_scale   consts.KsctlSpecialFlags = "autoscale"
	ksctl_feature_applications consts.KsctlSpecialFlags = "application"
)

func featureFlag(f *cobra.Command) {
	f.Flags().StringP("feature-flags", "", "", `Experimental Features: Supported values with comma seperated: [autoscale,application]`)
	// f.Flags().StringArrayP("feature-flags", "", nil, `Supported values: [autoscale]`)
}

func SetRequiredFeatureFlags(cmd *cobra.Command) {
	rawFeatures, err := cmd.Flags().GetString("feature-flags")
	if err != nil {
		return
	}
	features := strings.Split(rawFeatures, ",")

	for _, feature := range features {

		switch consts.KsctlSpecialFlags(feature) {
		case ksctl_feature_auto_scale:
			if err := os.Setenv(string(consts.KSCTL_FEATURE_FLAG_HA_AUTOSCALE), "true"); err != nil {
				if cli.Client.Storage != nil {
					cli.Client.Storage.Logger().Err("Unable to set the ha autoscale feature")
				} else {
					fmt.Println(errors.New("Unable to set the ha autoscale feature"))
				}
			}

		case ksctl_feature_applications:
			if err := os.Setenv(string(consts.KSCTL_FEATURE_FLAG_APPLICATIONS), "true"); err != nil {
				if cli.Client.Storage != nil {
					cli.Client.Storage.Logger().Err("Unable to set applications feature")
				} else {
					fmt.Println(errors.New("Unable to set the applications feature"))
				}
			}

		}
	}
}
