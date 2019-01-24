package aws

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	cf "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/stretchr/testify/assert"
)

var stack = NewStack(&fakeClient{})

// mock client
type fakeClient struct {
	cloudformationiface.CloudFormationAPI
}

func (fc *fakeClient) ListStacks(input *cf.ListStacksInput) (*cf.ListStacksOutput, error) {
	if *input.StackStatusFilter[0] == "UPDATE_COMPLETE" {
		return &cf.ListStacksOutput{
			StackSummaries: []*cf.StackSummary{
				&cf.StackSummary{
					StackName:   aws.String("test2"),
					StackStatus: aws.String("UPDATE_COMPLETE"),
				},
			},
		}, nil

	}

	if *input.StackStatusFilter[0] == "ERROR" {
		return nil, errors.New("error")
	}

	return &cf.ListStacksOutput{
		StackSummaries: []*cf.StackSummary{
			&cf.StackSummary{
				StackName:   aws.String("test1"),
				StackStatus: aws.String("DELETE_COMPLETE"),
			},
		},
	}, nil
}

func (fc *fakeClient) ValidateTemplate(input *cf.ValidateTemplateInput) (*cf.ValidateTemplateOutput, error) {
	return new(cf.ValidateTemplateOutput).SetDescription("testing"), nil
}

func (fc *fakeClient) CreateStack(input *cf.CreateStackInput) (*cf.CreateStackOutput, error) {
	return new(cf.CreateStackOutput).SetStackId("testing"), nil
}

func (fc *fakeClient) DeleteStack(input *cf.DeleteStackInput) (*cf.DeleteStackOutput, error) {
	return new(cf.DeleteStackOutput), nil
}

func (fc *fakeClient) DescribeStacks(input *cf.DescribeStacksInput) (*cf.DescribeStacksOutput, error) {
	var stacks []*cf.Stack

	sampleStack := new(cf.Stack).
		SetStackName("test").
		SetStackStatus(cf.StackStatusCreateComplete)

	stacks = append(stacks, sampleStack)

	return &cf.DescribeStacksOutput{
		Stacks: stacks,
	}, nil
}

func (fc *fakeClient) DescribeStackEvents(input *cf.DescribeStackEventsInput) (*cf.DescribeStackEventsOutput, error) {
	var events []*cf.StackEvent

	e := new(cf.StackEvent).
		SetEventId("test-event").
		SetStackId("test-stack-id").
		SetStackName("test")

	events = append(events, e)

	return &cf.DescribeStackEventsOutput{
		StackEvents: events,
	}, nil
}

func (fc *fakeClient) DetectStackDrift(input *cf.DetectStackDriftInput) (*cf.DetectStackDriftOutput, error) {
	return new(cf.DetectStackDriftOutput).SetStackDriftDetectionId("detect-id-123abc"), nil
}

func (fc *fakeClient) DescribeStackResourceDrifts(input *cf.DescribeStackResourceDriftsInput) (*cf.DescribeStackResourceDriftsOutput, error) {
	var drifts []*cf.StackResourceDrift

	d := new(cf.StackResourceDrift).
		SetStackId("test-abc").
		SetStackResourceDriftStatus(cf.StackDriftStatusDrifted)

	drifts = append(drifts, d)

	return &cf.DescribeStackResourceDriftsOutput{
		StackResourceDrifts: drifts,
	}, nil
}

func (fc *fakeClient) DescribeStackDriftDetectionStatus(input *cf.DescribeStackDriftDetectionStatusInput) (*cf.DescribeStackDriftDetectionStatusOutput, error) {
	return new(cf.DescribeStackDriftDetectionStatusOutput).
		SetDetectionStatus(cf.StackDriftDetectionStatusDetectionComplete).
		SetStackDriftDetectionId("abc-test").
		SetStackDriftStatus(cf.StackDriftStatusDrifted).
		SetStackId("test"), nil
}

func TestTagSlice(t *testing.T) {
	data := map[string]string{
		"Name": "testing",
	}

	tags := NewStack(&fakeClient{}).TagSlice(data)
	assert.Equal(t, 1, len(tags))
	assert.Equal(t, "testing", *tags[0].Value)
}

func TestParamSlice(t *testing.T) {
	data := map[string]string{
		"S3Name": "testing",
	}

	params := NewStack(&fakeClient{}).ParamSlice(data)
	assert.Equal(t, 1, len(params))
	assert.Equal(t, "testing", *params[0].ParameterValue)
}

func TestListStacks(t *testing.T) {
	testData := []map[string]string{
		nil,
		map[string]string{"filter": "UPDATE_COMPLETE"},
		map[string]string{"filter": "ERROR"},
	}

	for _, td := range testData {
		stackSum, err := stack.ListStacks("", td["filter"])
		if td["filter"] == "ERROR" {
			assert.Error(t, err)
		} else if td["filter"] == "UPDATE_COMPLETE" {
			assert.NoError(t, err)
			assert.Equal(t, "test2", *stackSum[0].StackName)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, 1, len(stackSum))
		}
	}
}

func TestValidateTemplate(t *testing.T) {
	var tpl []byte
	var url string

	// test no params
	_, err := stack.ValidateTemplate(tpl, url)
	assert.Error(t, err)

	// test url
	_, err = stack.ValidateTemplate(tpl, "https://s3")
	assert.NoError(t, err)
}

func TestCreateStack(t *testing.T) {
	_, err := stack.CreateStack("testing", nil, nil, nil, "https://s3")
	assert.NoError(t, err)
}

func TestDeleteStack(t *testing.T) {
	_, err := stack.DeleteStack("testing")
	assert.NoError(t, err)
}

func TestDescribeStack(t *testing.T) {
	s, err := stack.DescribeStack("")
	assert.Error(t, err)

	s, err = stack.DescribeStack("test")
	assert.NoError(t, err)
	assert.True(t, len(s) > 0)
}

func TestDescribeStacks(t *testing.T) {
	s, err := stack.DescribeStacks()
	assert.NoError(t, err)
	assert.True(t, len(s) > 0)
}

func TestDescribeStackEvents(t *testing.T) {
	se, err := stack.DescribeStackEvents("test")
	assert.NoError(t, err)
	assert.True(t, len(se) > 0)
}

func TestDetectStackDrift(t *testing.T) {
	id, err := stack.DetectStackDrift("test")
	assert.NoError(t, err)
	assert.True(t, len(id) > 0)
}

func TestDescribeStackResourceDrifts(t *testing.T) {
	out, err := stack.DescribeStackResourceDrifts("test", cf.StackDriftStatusDrifted)
	assert.NoError(t, err)
	assert.True(t, len(out) > 0)
}

func TestDescribeStackDriftDetectionStatus(t *testing.T) {
	out, err := stack.DescribeStackDriftDetectionStatus("test")
	assert.NoError(t, err)
	assert.IsType(t, new(cf.DescribeStackDriftDetectionStatusOutput), out)
}
