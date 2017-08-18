// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package resync ties datasync into the lifecycle of CN-Infra apps. In
// particular the package implements:
//  1. Data resynchronization (resync) event after the start/restart of
//     the CN-Infra app or when ther app's connectivity to external data
//     sources/sinks is lost and subsequently restored;
//  2. Data resynchronization of the datasync transport upon Init()
//     and Close()
package resync
