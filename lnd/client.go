package lnd

import (
	"github.com/lightningnetwork/lnd/macaroons"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	macaroon "gopkg.in/macaroon.v2"
	"io/ioutil"
)

// Client holds the LND client connection.
type Client struct {
	Conn *grpc.ClientConn
}

// NewClient returns a new Client.
func NewClient() *Client {
	c := &Client{}
	return c
}

// Connect connects the Client to the LND at the provided address and authenticates using
// the TLS cert found at tlsCertPath and the admin macaroon at adminMacaroonPath.
func (c *Client) Connect(address string, tlsCertPath string, adminMacaroonPath string) error {
	var opts []grpc.DialOption

	creds, err := credentials.NewClientTLSFromFile(tlsCertPath, "")
	if err != nil {
		return err
	}
	opts = append(opts, grpc.WithTransportCredentials(creds))

	macaroonBytes, err := ioutil.ReadFile(adminMacaroonPath)
	if err != nil {
		return err
	}
	mac := &macaroon.Macaroon{}
	if err = mac.UnmarshalBinary(macaroonBytes); err != nil {
		return err
	}
	macCreds, err := macaroons.NewMacaroonCredential(mac)
	if err != nil {
		return err
	}
	opts = append(opts, grpc.WithPerRPCCredentials(macCreds))

	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		return err
	}
	c.Conn = conn
	return nil
}

// IsConnected returns a boolean: whether the Client is connected to LND.
func (c *Client) IsConnected() bool {
	return c.Conn != nil
}
