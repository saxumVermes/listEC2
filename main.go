package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const (
	PENDING = iota * 16
	RUNNING
	SHUTTING_DOWN
	TERMINATED
	STOPPING
	STOPPED
)

type SlackRequest struct {
	token       string
	teamId      string
	teamDomain  string
	channelId   string
	channelName string
	command     string
	userId      string
	userName    string
	text        string
	responseUrl string
	triggerId   string
}

func parseRequest(r *http.Request) (s SlackRequest) {

	s.token = r.FormValue("token")
	s.teamId = r.FormValue("team_id")
	s.command = r.FormValue("command")
	s.userId = r.FormValue("user_id")
	s.userName = r.FormValue("user_name")
	s.text = r.FormValue("text")

	return s
}

func status(status string) (int, error) {
	switch status {
	case "pending":
		return PENDING, nil
	case "running":
		return RUNNING, nil
		type listB map[string]interface{}
	case "shutting-down":
		return SHUTTING_DOWN, nil
	case "terminated":
		return TERMINATED, nil
	case "stopping":
		return STOPPING, nil
	case "stopped":
		return STOPPED, nil
	case "all":
		return 1, nil
	}
	return 0, fmt.Errorf("Wrong input try with these: pending|running|shutting-down|terminated|stopping|stopped")
}

func handleListEC2Instances(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	sReq := parseRequest(r)
	s, err := status(sReq.text)
	if err != nil {
		fmt.Fprintf(w, "%v", err)
		return
	}
	r.Header.Set("Content-Type", "application/json")
	svc := ec2.New(session.New())
	filter := &ec2.DescribeInstancesInput{}

	if s != 1 {
		filter.Filters = []*ec2.Filter{
			{
				Name: aws.String("instance-state-code"),
				Values: []*string{
					aws.String(strconv.Itoa(s)),
				},
			},
		}
	}

	result, err := svc.DescribeInstances(filter)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return
	}

	var statusList = map[string]int{}
	for _, res := range result.Reservations {
		for _, inst := range res.Instances {
			statusList[*inst.State.Name]++
		}
	}

	for i, v := range statusList {
		fmt.Fprintf(w, "%-10s %d\n", i, v)
	}
}

func main() {
	http.HandleFunc("/list-ec2", handleListEC2Instances)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
