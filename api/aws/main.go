/*
Kubesimplify
@maintainer:
*/

package eks

import (
	"fmt"
	"go/token"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	util "github.com/kubesimplify/ksctl/api/utils"
)

type AWSCLUSTER struct {
	ClusterName  string
	NodeType     string
	NodeCount    int64
	InstanceType string // like t2.micro,ec2.large etc
	Config       string
	Token        string // token is used for authentication
}

// fetchAPIKey returns the API key from the cred/civo file store
func fetchAPIKey() string {

	_, err := os.ReadFile(util.GetPath(0, "aws"))
	if err != nil {
		return ""
	}

	return ""
}

func Credentials() bool {
	// _, err := os.OpenFile(util.GetPath(0, "aws"), os.O_WRONLY, 0640)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return false
	// }
	acckey := ""
	secacckey := ""
	func() {
		fmt.Println("Enter your ACCESS-KEY: ")
		_, err := fmt.Scan(&acckey)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println("Enter your SECRET-ACCESS-KEY: ")
		_, err = fmt.Scan(&secacckey)
		if err != nil {
			panic(err.Error())
		}
	}()

	// _, err = file.Write([]byte(fmt.Sprintf(`Access-Key: %s
	// 	Secret-Access-Key: %s`, acckey, secacckey)))
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return false
	// }
	return true
}

func (obj *AWSCLUSTER) Authentication(token string) bool {
	CreateAWSClient(token, "us-east-1")

	obj.Token = token
	return true
}

func CreateAWSClient(token string, Region string) (*sts.STS, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(Region),
	})
	if err != nil {
		return nil, err
	}
	svc := sts.New(sess)

	return svc, nil
}
