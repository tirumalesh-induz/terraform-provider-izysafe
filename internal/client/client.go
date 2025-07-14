package client

import (
	"context"
	"encoding/base64"
	"fmt"

	"crypto/tls"
	"crypto/x509"
	"net/url"
	"sync"
	"time"

	"terraform-provider-izysafe/internal/proto/request"
	"terraform-provider-izysafe/internal/proto/response"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

var (
	globalClient *Client
	clientMu     sync.Mutex
)

func GetClient(token, endpoint, pin string) *Client {
	clientMu.Lock()
	defer clientMu.Unlock()
	if globalClient == nil {
		globalClient, _ = New(context.Background(), endpoint, token, pin)
	}
	return globalClient
}

type Client struct {
	mu    sync.Mutex
	conn  *websocket.Conn
	token string
	pin   string
	Email string
}

var caCertBytes = []byte(`-----BEGIN CERTIFICATE-----
MIICjTCCAhSgAwIBAgIIdebfy8FoW6gwCgYIKoZIzj0EAwIwfDELMAkGA1UEBhMC
VVMxDjAMBgNVBAgMBVRleGFzMRAwDgYDVQQHDAdIb3VzdG9uMRgwFgYDVQQKDA9T
U0wgQ29ycG9yYXRpb24xMTAvBgNVBAMMKFNTTC5jb20gUm9vdCBDZXJ0aWZpY2F0
aW9uIEF1dGhvcml0eSBFQ0MwHhcNMTYwMjEyMTgxNDAzWhcNNDEwMjEyMTgxNDAz
WjB8MQswCQYDVQQGEwJVUzEOMAwGA1UECAwFVGV4YXMxEDAOBgNVBAcMB0hvdXN0
b24xGDAWBgNVBAoMD1NTTCBDb3Jwb3JhdGlvbjExMC8GA1UEAwwoU1NMLmNvbSBS
b290IENlcnRpZmljYXRpb24gQXV0aG9yaXR5IEVDQzB2MBAGByqGSM49AgEGBSuB
BAAiA2IABEVuqVDEpiM2nl8ojRfLliJkP9x6jh3MCLOicSS6jkm5BBtHllirLZXI
7Z4INcgn64mMU1jrYor+8FsPazFSY0E7ic3s7LaNGdM0B9y7xgZ/wkWV7Mt/qCPg
CemB+vNH06NjMGEwHQYDVR0OBBYEFILRhXMw5zUE044CkvvlpNHEIejNMA8GA1Ud
EwEB/wQFMAMBAf8wHwYDVR0jBBgwFoAUgtGFczDnNQTTjgKS++Wk0cQh6M0wDgYD
VR0PAQH/BAQDAgGGMAoGCCqGSM49BAMCA2cAMGQCMG/n61kRpGDPYbCWe+0F+S8T
kdzt5fxQaxFGRrMcIQBiu77D5+jNB5n5DQtdcj7EqgIwH7y6C+IwJPt8bYBVCpk+
gA0z5Wajs6O7pdWLjwkspl1+4vAHCGht0nxpbl/f5Wpl
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIDejCCAv+gAwIBAgIQHNcSEt4VENkSgtozEEoQLzAKBggqhkjOPQQDAzB8MQsw
CQYDVQQGEwJVUzEOMAwGA1UECAwFVGV4YXMxEDAOBgNVBAcMB0hvdXN0b24xGDAW
BgNVBAoMD1NTTCBDb3Jwb3JhdGlvbjExMC8GA1UEAwwoU1NMLmNvbSBSb290IENl
cnRpZmljYXRpb24gQXV0aG9yaXR5IEVDQzAeFw0xOTAzMDcxOTQyNDJaFw0zNDAz
MDMxOTQyNDJaMG8xCzAJBgNVBAYTAlVTMQ4wDAYDVQQIDAVUZXhhczEQMA4GA1UE
BwwHSG91c3RvbjERMA8GA1UECgwIU1NMIENvcnAxKzApBgNVBAMMIlNTTC5jb20g
U1NMIEludGVybWVkaWF0ZSBDQSBFQ0MgUjIwdjAQBgcqhkjOPQIBBgUrgQQAIgNi
AASEOWn30uEYKDLFu4sCjFQ1VupFaeMtQjqVWyWSA7+KFljnsVaFQ2hgs4cQk1f/
RQ2INSwdVCYU0i5qsbom20rigUhDh9dM/r6bEZ75eFE899kSCI14xqThYVLPdLEl
+dyjggFRMIIBTTASBgNVHRMBAf8ECDAGAQH/AgEAMB8GA1UdIwQYMBaAFILRhXMw
5zUE044CkvvlpNHEIejNMHgGCCsGAQUFBwEBBGwwajBGBggrBgEFBQcwAoY6aHR0
cDovL3d3dy5zc2wuY29tL3JlcG9zaXRvcnkvU1NMY29tLVJvb3RDQS1FQ0MtMzg0
LVIxLmNydDAgBggrBgEFBQcwAYYUaHR0cDovL29jc3BzLnNzbC5jb20wEQYDVR0g
BAowCDAGBgRVHSAAMB0GA1UdJQQWMBQGCCsGAQUFBwMCBggrBgEFBQcDATA7BgNV
HR8ENDAyMDCgLqAshipodHRwOi8vY3Jscy5zc2wuY29tL3NzbC5jb20tZWNjLVJv
b3RDQS5jcmwwHQYDVR0OBBYEFA10Zgpen+Is7NXCXSUEf3Uyuv99MA4GA1UdDwEB
/wQEAwIBhjAKBggqhkjOPQQDAwNpADBmAjEAxYt6Ylk/N8Fch/3fgKYKwI5A011Q
MKW0h3F9JW/NX/F7oYtWrxljheH8n2BrkDybAjEAlCxkLE0vQTYcFzrR24oogyw6
VkgTm92+jiqJTO5SSA9QUa092S5cTKiHkH2cOM6m
-----END CERTIFICATE-----`)

func New(ctx context.Context, endpoint, token string, pin string) (*Client, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCertBytes); !ok {
		return nil, fmt.Errorf("failed to load certificates")
	}
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		TLSClientConfig:  tlsConfig,
	}

	conn, _, err := dialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return nil, err
	}

	data, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("failed to hex decode token: %w", err)
	}
	signin := request.SignIn{
		Data: data,
		Pin:  &pin,
	}
	req := request.Request{
		Operation: &request.Request_SignIn{
			SignIn: &signin,
		},
	}
	encReq, err := proto.Marshal(&req)
	if err != nil {
		return nil, err
	}
	if err := conn.WriteMessage(websocket.BinaryMessage, encReq); err != nil {
		return nil, err
	}

	var responseObj response.Response
	_, resp, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(resp, &responseObj); err != nil {
		return nil, err

	}
	if responseObj.GetSignIn().Status != response.Status_SUCCESS {
		return nil, err
	}
	Email := responseObj.GetSignIn().Email
	client := &Client{
		mu:    sync.Mutex{},
		conn:  conn,
		Email: Email,
		pin:   pin,
		token: token,
	}
	return client, nil
}

func (c *Client) Send(req *request.Request) (*response.Response, error) {
	clientMu.Lock()
	defer clientMu.Unlock()
	if c == nil {
		return nil, nil
	}
	protoReq, _ := proto.Marshal(req)
	if c.conn == nil {
		return nil, fmt.Errorf("connection is nil")
	}
	err := c.conn.WriteMessage(websocket.BinaryMessage, protoReq)
	if err != nil {
		return nil, err
	}
	_, resp, err := c.conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	var responseObj response.Response
	if err := proto.Unmarshal(resp, &responseObj); err != nil {
		return nil, err
	}
	if proto.Equal(&responseObj, &response.Response{}) {
		return nil, fmt.Errorf("empty response")
	}
	return &responseObj, nil
}
