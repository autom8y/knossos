package slack

import (
	"context"
	"fmt"
	"testing"

	slackapi "github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testSlackAPI implements SlackAPI for unit tests.
type testSlackAPI struct {
	postMessageCalls  []postMessageCall
	authTestErr       error
	postMessageErr    error
	updateMessageErr  error
}

type postMessageCall struct {
	channelID string
}

func (t *testSlackAPI) PostMessage(channelID string, options ...slackapi.MsgOption) (string, string, error) {
	t.postMessageCalls = append(t.postMessageCalls, postMessageCall{channelID: channelID})
	return "C001", "1234567890.123456", t.postMessageErr
}

func (t *testSlackAPI) UpdateMessage(channelID, timestamp string, options ...slackapi.MsgOption) (string, string, string, error) {
	return "C001", "1234567890.123456", "", t.updateMessageErr
}

func (t *testSlackAPI) AuthTest() (*slackapi.AuthTestResponse, error) {
	if t.authTestErr != nil {
		return nil, t.authTestErr
	}
	return &slackapi.AuthTestResponse{
		UserID: "U001",
		TeamID: "T001",
	}, nil
}

func TestSlackClient_SendBlocks(t *testing.T) {
	api := &testSlackAPI{}
	client := NewSlackClientWithAPI(api, "xoxb-test")

	blocks := []slackapi.Block{
		slackapi.NewSectionBlock(
			slackapi.NewTextBlockObject(slackapi.MarkdownType, "Hello", false, false),
			nil, nil,
		),
	}

	err := client.SendBlocks("C001", "1234567890.123456", blocks)
	require.NoError(t, err)
	require.Len(t, api.postMessageCalls, 1)
	assert.Equal(t, "C001", api.postMessageCalls[0].channelID)
}

func TestSlackClient_SendBlocks_Error(t *testing.T) {
	api := &testSlackAPI{
		postMessageErr: fmt.Errorf("channel_not_found"),
	}
	client := NewSlackClientWithAPI(api, "xoxb-test")

	err := client.SendBlocks("C001", "", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "send blocks to C001")
}

func TestSlackClient_SendBlocks_NoThread(t *testing.T) {
	api := &testSlackAPI{}
	client := NewSlackClientWithAPI(api, "xoxb-test")

	err := client.SendBlocks("C001", "", nil)
	require.NoError(t, err)
	require.Len(t, api.postMessageCalls, 1)
}

func TestSlackClient_HealthCheck_Success(t *testing.T) {
	api := &testSlackAPI{}
	client := NewSlackClientWithAPI(api, "xoxb-test")

	err := client.HealthCheck(context.Background())
	require.NoError(t, err)
}

func TestSlackClient_HealthCheck_Error(t *testing.T) {
	api := &testSlackAPI{
		authTestErr: fmt.Errorf("invalid_auth"),
	}
	client := NewSlackClientWithAPI(api, "xoxb-test")

	err := client.HealthCheck(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "slack auth.test failed")
}
