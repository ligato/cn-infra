//  Copyright (c) 2018 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package greeter

import (
	"context"
	"errors"

	"google.golang.org/grpc/examples/helloworld/helloworld"
)

// Service is an empty struct to hang SayHello off of
type Service struct{}

// SayHello takes a context and a HelloRequest and returns a HelloReply
func (*Service) SayHello(ctx context.Context, request *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	if request.Name == "" {
		return nil, errors.New("not filled name in the request")
	}
	return &helloworld.HelloReply{Message: "Greetings " + request.Name}, nil
}
