package main

import (
	"flag"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"

	"go.ligato.io/cn-infra/v2/examples/grpc-plugin/insecure"
	"go.ligato.io/cn-infra/v2/logging/logs"
)

const (
	defaultAddress = "localhost:9111"
	defaultName    = "world"
)

var address = defaultAddress
var name = defaultName

func main() {
	flag.StringVar(&address, "address", defaultAddress, "address of GRPC server")
	flag.StringVar(&name, "name", defaultName, "name used in GRPC request")
	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.Dial(address,
		//grpc.WithInsecure(),
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(insecure.CertPool, "")),
		grpc.WithPerRPCCredentials(tokenAuth{
			token: "testtoken",
		}),
	)

	if err != nil {
		logs.DefaultLogger().Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	r, err := c.SayHello(context.Background(), &pb.HelloRequest{Name: name})
	if err != nil {
		logs.DefaultLogger().Fatalf("could not greet: %v", err)
	}
	logs.DefaultLogger().Printf("Reply: %s (received from server)", r.Message)
}

type tokenAuth struct {
	token string
}

func (t tokenAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + t.token,
	}, nil
}

func (tokenAuth) RequireTransportSecurity() bool {
	return true
}
