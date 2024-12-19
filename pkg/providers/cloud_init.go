package providers

import "fmt"

func CloudInitScript(resName string) (string, error) {

	postfixStr, err := GenRandomString(5)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`#!/bin/bash
sudo hostname %s-%s

sudo cp /etc/localtime /etc/localtime.backup

sudo ln -sf /usr/share/zoneinfo/UTC /etc/localtime

`, resName, postfixStr), nil
}
