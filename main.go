package main

import (
	"flag"
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
	text string
}

func parseRequest(r *http.Request) (s SlackRequest) {
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
	input := &ec2.DescribeInstancesInput{}

	if s != 1 {
		input.Filters = []*ec2.Filter{
			{
				Name: aws.String("instance-state-code"),
				Values: []*string{
					aws.String(strconv.Itoa(s)),
				},
			},
		}
	}

	result, err := svc.DescribeInstances(input)
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
	port := flag.String("http", "8080", "Specify listening port")
	flag.Parse()

	http.HandleFunc("/list-ec2", handleListEC2Instances)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *port), nil))
}
