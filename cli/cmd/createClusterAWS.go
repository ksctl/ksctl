package cmd

/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/
import (
	"github.com/aws/aws-sdk-go"
	"github.com/kubesimplify/ksctl/api/eks"
	"github.com/spf13/cobra"
)


var createClusterAws = &cobra.Command{
	Use:   "aws",
	Short: "Use to create a EKS cluster in AWS",
	Long: `It is used to create cluster with the given name from user. For example:

ksctl create-cluster aws <arguments to civo cloud provider>
`,
	Run: func(cmd *cobra.Command, args []string) {

		svc := eks.New




	},
}


var (

	awclientToken       string
	awclusterName       string
	awsecuritygrpID     []string
	awsubnetID			[]string
	awroleARN			string
	awVersion			string
)




func init() {
	createClusterCmd.AddCommand(createClusterAws)
}
