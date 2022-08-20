package cfg

// Lnd holds options related to connecting to the backend LND instance.
type Lnd struct {
	Address       string `long:"address" description:"The address of the backend LND gRPC server"`
	TlsCert       string `long:"tlscert" description:"The LND TLS cert file"`
	AdminMacaroon string `long:"adminmacaroon" description:"The LND admin macaroon file"`
}

// Tls holds options related to TLS serving of the REST and gRPC APIs.
type Tls struct {
	EnableTls         bool   `long:"enable" description:"Enable TLS"`
	CertPath          string `long:"cert" description:"Path to TLS cert file"`
	KeyPath           string `long:"key" description:"Path to TLS key file"`
	CertValidityHours int64  `long:"certValidityHours" description:"Validity duration of the ssl cert"`
}

// Serving holds options related to the serving of the REST and gRPC APIs.
type Serving struct {
	Hostname string `long:"host" description:"Host to serve from"`
	Rest     bool   `long:"rest" description:"Enable serving the REST API"`
	RestPort uint16 `long:"reston" description:"The port to serve REST on"`
	Grpc     bool   `long:"grpc" description:"Enable the gRPC server"`
	GrpcPort uint16 `long:"grpcon" description:"The port to serve gRPC on"`
	Tls      *Tls
}
