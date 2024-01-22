package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// helper recieved from https://ubuntu.com/tutorials/search-and-launch-ubuntu-22-04-in-aws-using-cli#2-search-for-the-right-ami
func trustedSource(id string) bool {
	// 679593333241
	// 099720109477
	if strings.Compare(id, "679593333241") != 0 && strings.Compare(id, "099720109477") != 0 {
		return false
	}
	return true
}

func getLatestAmazonLinux2AMI(client *ec2.Client) (string, error) {
	// Specify the filter for Amazon Linux 2 images
	imageFilter := &ec2.DescribeImagesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("name"),
				Values: []string{"ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server*"},
			},
			{
				Name:   aws.String("architecture"),
				Values: []string{"x86_64"},
			},
			{
				Name:   aws.String("owner-alias"),
				Values: []string{"amazon"},
			},
		},
	}

	// Get the latest Amazon Linux 2 AMI
	resp, err := client.DescribeImages(context.TODO(), imageFilter)
	if err != nil {
		return "", fmt.Errorf("failed to describe images: %w", err)
	}

	if len(resp.Images) == 0 {
		return "", fmt.Errorf("no images found")
	}

	var savedImages []types.Image

	for _, i := range resp.Images {
		if trustedSource(*i.OwnerId) && *i.Public {
			savedImages = append(savedImages, i)
		}
	}
	sort.Slice(savedImages, func(i, j int) bool {
		return *savedImages[i].CreationDate > *savedImages[j].CreationDate
	})

	for x := 0; x < 2; x++ {
		i := savedImages[x]
		fmt.Println("=======")
		if i.ImageOwnerAlias != nil {
			fmt.Printf("%#+v\n", *i.ImageOwnerAlias)
		}
		fmt.Printf("%#v\n", *i.CreationDate)
		fmt.Printf("%#v\n", *i.Public)
		fmt.Printf("%#v\n", *i.OwnerId)
		fmt.Printf("%#v\n", i.Architecture.Values())
		fmt.Printf("%#v\n", *i.Name)
		fmt.Printf("%#v\n", *i.ImageId)
		fmt.Println("=======")
	}

	// Get the latest image ID
	latestAMI := savedImages[0].ImageId
	return *latestAMI, nil
}
