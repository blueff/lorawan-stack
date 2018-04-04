// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"fmt"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
)

func defaultGateway(in interface{}) (*ttnpb.Gateway, error) {
	if gtw, ok := in.(store.Gateway); ok {
		return gtw.GetGateway(), nil
	}

	if gtw, ok := in.(ttnpb.Gateway); ok {
		return &gtw, nil
	}

	if ptr, ok := in.(*ttnpb.Gateway); ok {
		return ptr, nil
	}

	return nil, fmt.Errorf("Expected: '%v' to be of type ttnpb.Gateway but it was not", in)
}

// ShouldBeGateway checks if two Gateways resemble each other.
func ShouldBeGateway(actual interface{}, expected ...interface{}) string {
	if len(expected) != 1 {
		return fmt.Sprintf("Expected: one gateway to match but got %v", len(expected))
	}

	a, s := defaultGateway(actual)
	if s != nil {
		return s.Error()
	}

	b, s := defaultGateway(expected[0])
	if s != nil {
		return s.Error()
	}

	return all(
		ShouldBeGatewayIgnoringAutoFields(a, b),
		assertions.ShouldHappenWithin(a.UpdatedAt, time.Millisecond, b.UpdatedAt),
		assertions.ShouldHappenWithin(a.CreatedAt, time.Millisecond, b.CreatedAt),
	)
}

// ShouldBeGatewayIgnoringAutoFields checks if two Gateways resemble each other
// without looking at fields that are generated by the database: created.
func ShouldBeGatewayIgnoringAutoFields(actual interface{}, expected ...interface{}) string {
	if len(expected) != 1 {
		return fmt.Sprintf("Expected: one gateway to match but got %v", len(expected))
	}

	a, s := defaultGateway(actual)
	if s != nil {
		return s.Error()
	}

	b, s := defaultGateway(expected[0])
	if s != nil {
		return s.Error()
	}

	return all(
		assertions.ShouldEqual(a.GatewayID, b.GatewayID),
		assertions.ShouldEqual(a.Description, b.Description),
		assertions.ShouldEqual(a.FrequencyPlanID, b.FrequencyPlanID),
		assertions.ShouldEqual(a.ClusterAddress, b.ClusterAddress),
		assertions.ShouldResemble(a.Antennas, b.Antennas),
		assertions.ShouldResemble(a.Radios, b.Radios),
		assertions.ShouldBeTrue(a.ActivatedAt.Equal(b.ActivatedAt)),
		assertions.ShouldResemble(a.PrivacySettings, b.PrivacySettings),
		assertions.ShouldEqual(a.AutoUpdate, b.AutoUpdate),
		assertions.ShouldResemble(a.Platform, b.Platform),
		assertions.ShouldResemble(a.Attributes, b.Attributes),
	)
}

func gatewayAntenna(in interface{}) (*ttnpb.GatewayAntenna, error) {
	if antenna, ok := in.(*ttnpb.GatewayAntenna); ok {
		return antenna, nil
	}

	if antenna, ok := in.(ttnpb.GatewayAntenna); ok {
		return &antenna, nil
	}

	return nil, fmt.Errorf("Expected: '%v' to be of type ttnpb.GatewayAntenna but it was not", in)
}

// ShouldBeGatewayAntenna checks if two Gateway Antennas resemble each other.
func ShouldBeGatewayAntenna(actual interface{}, expected ...interface{}) string {
	if len(expected) != 1 {
		return fmt.Sprintf("Expected: one gateway antenna to match but got %v", len(expected))
	}

	a, s := gatewayAntenna(actual)
	if s != nil {
		return s.Error()
	}

	b, s := gatewayAntenna(expected[0])
	if s != nil {
		return s.Error()
	}

	return all(
		assertions.ShouldEqual(a.Gain, b.Gain),
		assertions.ShouldResemble(a.Location, b.Location),
		assertions.ShouldEqual(a.Type, b.Type),
		assertions.ShouldEqual(a.Model, b.Model),
		assertions.ShouldEqual(a.Placement, b.Placement),
	)
}
