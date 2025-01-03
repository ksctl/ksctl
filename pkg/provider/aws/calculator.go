package aws

//
//import (
//	"context"
//	"errors"
//	"github.com/aws/aws-sdk-go-v2/aws"
//	"github.com/aws/aws-sdk-go-v2/service/bcmpricingcalculator"
//	"github.com/aws/aws-sdk-go-v2/service/bcmpricingcalculator/types"
//	"github.com/gookit/goutil/dump"
//	"log"
//	"math/rand"
//	"time"
//)
//
//// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/bcmpricingcalculator#Client.BatchCreateBillScenarioUsageModification
//// https://docs.aws.amazon.com/aws-cost-management/latest/APIReference/API_Types_AWS_Billing_and_Cost_Management_Pricing_Calculator.html
//
//func Calc(cfg aws.Config) error {
//	svc := bcmpricingcalculator.NewFromConfig(cfg)
//
//	accId, err := getAccountId(cfg)
//	if err != nil {
//		return err
//	}
//	// Create a new Bill Scenario
//	createBillScenarioInput := &bcmpricingcalculator.CreateBillScenarioInput{
//		Name: aws.String("MyBillScenario"),
//	}
//	createBillScenarioOutput, err := svc.CreateBillScenario(context.TODO(), createBillScenarioInput)
//	if err != nil {
//		log.Fatalf("unable to create bill scenario, %v", err)
//	}
//
//	if errors.As(err, &types.ConflictException{}) {
//		_, err := svc.DeleteBillEstimate(context.TODO(), createBillScenarioInput)
//		if err != nil {
//			log.Fatalf("unable to create bill scenario, %v", err)
//		}
//	}
//	billScenarioId := createBillScenarioOutput.Id
//	inp := &bcmpricingcalculator.BatchCreateBillScenarioUsageModificationInput{
//		BillScenarioId: billScenarioId,
//		UsageModifications: []types.BatchCreateBillScenarioUsageModificationEntry{
//			{
//				Key:            aws.String("ec2"),
//				Operation:      aws.String("RunInstances"),
//				ServiceCode:    aws.String("AmazonEC2"),
//				UsageAccountId: accId,
//				UsageType:      aws.String("BoxUsage:t3.micro"),
//			},
//		},
//	}
//
//	//c, err := svc.BatchCreateBillScenarioUsageModification(
//	//	context.TODO(),
//	//	inp,
//	//)
//	//if err != nil {
//	//	return err
//	//}
//	//
//	//ddd := dump.NewWithOptions(dump.SkipNilField(), dump.SkipPrivate())
//	//
//	//for _, v := range c.Items {
//	//	ddd.Println(v)
//	//}
//	//return nil
//	// Implement retry logic with exponential backoff
//	maxAttempts := 3
//	for attempt := 1; attempt <= maxAttempts; attempt++ {
//		c, err := svc.BatchCreateBillScenarioUsageModification(context.TODO(), inp)
//		if err == nil {
//			ddd := dump.NewWithOptions(dump.SkipNilField(), dump.SkipPrivate())
//			for _, v := range c.Items {
//				ddd.Println(v)
//			}
//			return nil
//		}
//
//		var apiErr *types.InternalServerException
//		if errors.As(err, &apiErr) {
//			if attempt < maxAttempts {
//				backoff := time.Duration(rand.Intn(1<<attempt)) * time.Second
//				time.Sleep(backoff)
//				continue
//			}
//		}
//		return err
//	}
//	return errors.New("exceeded maximum number of attempts")
//}
