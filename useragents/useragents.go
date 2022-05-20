package useragents

import (
	"bufio"
	"errors"
	"math/rand"
	"os"
	"time"
)

type UserAgentClient struct {
	agentList []string
}

func (u *UserAgentClient) LoadUserAgents(fp string) error {
	var err error
	if fp == "" {
		return errors.New("no filename provided")
	}
	file, err := os.Open(fp)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		u.agentList = append(u.agentList, scanner.Text())
	}
	return scanner.Err()

}

func (u *UserAgentClient) GetRandomUserAgent() (string, error) {

	if u.agentList == nil {
		return "", errors.New("agent list is nil")
	}
	rand.Seed(time.Now().UnixNano())

	i := rand.Intn(len(u.agentList)-0) + 0

	return u.agentList[i], nil

}
