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

package identityserver

import "github.com/TheThingsNetwork/ttn/pkg/ttnpb"

// claims is the type that represents a claims to do something in the Identity Server.
type claims struct {
	EntityIdentifiers interface{}
	Source            string
	Rights            []ttnpb.Right
}

// UserIdentifiers returns the ttnpb.UserIdentifiers of the user this claims are
// related to, or a zero-valued ttnpb.UserIdentifiers otherwise.
func (c *claims) UserIdentifiers() (ids ttnpb.UserIdentifiers) {
	if i, ok := c.EntityIdentifiers.(ttnpb.UserIdentifiers); ok {
		ids = i
	}
	return
}

// ApplicationIdentifiers returns the ttnpb.ApplicationIdentifiers of the application
// this claims are related to, or a zero-valued ttnpb.ApplicationIdentifiers otherwise.
func (c *claims) ApplicationIdentifiers() (ids ttnpb.ApplicationIdentifiers) {
	if i, ok := c.EntityIdentifiers.(ttnpb.ApplicationIdentifiers); ok {
		ids = i
	}
	return
}

// GatewayIdentifiers returns the ttnpb.GatewayIdentifiers of the gateway
// this claims are related to, or a zero-valued ttnpb.GatewayIdentifiers otherwise.
func (c *claims) GatewayIdentifiers() (ids ttnpb.GatewayIdentifiers) {
	if i, ok := c.EntityIdentifiers.(ttnpb.GatewayIdentifiers); ok {
		ids = i
	}
	return
}

// OrganizationIdentifiers returns the ttnpb.OrganizationIdentifiers of the organization
// this claims are related to, or a zero-valued ttnpb.OrganiationIdentifiers otherwise.
func (c *claims) OrganizationIdentifiers() (ids ttnpb.OrganizationIdentifiers) {
	if i, ok := c.EntityIdentifiers.(ttnpb.OrganizationIdentifiers); ok {
		ids = i
	}
	return
}

// HasRights checks whether or not the provided rights are included in the claims.
// It will only return true if all the provided rights are included in the claims.
func (c *claims) HasRights(rights ...ttnpb.Right) bool {
	ok := true
	for _, right := range rights {
		ok = ok && c.hasRight(right)
	}

	return ok
}

// hasRight checks whether or not the right is included in this claims.
func (c *claims) hasRight(right ttnpb.Right) bool {
	for _, r := range c.Rights {
		if r == right {
			return true
		}
	}
	return false
}
