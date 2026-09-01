package rpc

import (
	"testing"

	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/common/nacos_server"
	"github.com/nacos-group/nacos-sdk-go/v2/common/remote/rpc/rpc_request"
	"github.com/nacos-group/nacos-sdk-go/v2/common/remote/rpc/rpc_response"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {

}

// mock403Conn rejects every request with ErrorResponse(403), counting calls.
type mock403Conn struct {
	mockConn
	calls int
}

func (m *mock403Conn) request(request rpc_request.IRequest, timeoutMills int64, client *RpcClient) (rpc_response.IResponse, error) {
	m.calls++
	return &rpc_response.ErrorResponse{Response: &rpc_response.Response{
		ErrorCode: constant.NO_RIGHT,
		Message:   "Invalid signature",
	}}, nil
}

// A 403 response must be returned directly and mark re-login on the nacos server,
// instead of being swallowed by the retry loop with the same stale token.
func TestRequest_NoRightReturnsDirectly(t *testing.T) {
	client := &RpcClient{nacosServer: &nacos_server.NacosServer{}}
	client.rpcClientStatus = RUNNING
	conn := &mock403Conn{}
	client.SetCurrentConnection(conn)

	resp, err := client.Request(rpc_request.NewConfigBatchListenRequest(1), 3000)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, constant.NO_RIGHT, resp.GetErrorCode())
	assert.Equal(t, 1, conn.calls, "403 must not be retried with the same stale token")
}
