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

package rights

import (
	"context"
	"sync"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func requireRights(ctx context.Context, id string) (res struct {
	AppErr error
	GtwErr error
	OrgErr error
}) {
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		res.AppErr = RequireApplication(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: id}, ttnpb.RIGHT_APPLICATION_INFO)
		wg.Done()
	}()
	go func() {
		res.GtwErr = RequireGateway(ctx, ttnpb.GatewayIdentifiers{GatewayID: id}, ttnpb.RIGHT_GATEWAY_INFO)
		wg.Done()
	}()
	go func() {
		res.OrgErr = RequireOrganization(ctx, ttnpb.OrganizationIdentifiers{OrganizationID: id}, ttnpb.RIGHT_ORGANIZATION_INFO)
		wg.Done()
	}()
	wg.Wait()
	return
}

func TestRequire(t *testing.T) {
	a := assertions.New(t)

	a.So(func() {
		RequireApplication(context.Background(), ttnpb.ApplicationIdentifiers{}, ttnpb.RIGHT_APPLICATION_INFO)
	}, should.Panic)
	a.So(func() {
		RequireGateway(context.Background(), ttnpb.GatewayIdentifiers{}, ttnpb.RIGHT_GATEWAY_INFO)
	}, should.Panic)
	a.So(func() {
		RequireOrganization(context.Background(), ttnpb.OrganizationIdentifiers{}, ttnpb.RIGHT_ORGANIZATION_INFO)
	}, should.Panic)

	fooCtx := newContext(context.Background(), Rights{
		ApplicationRights: map[ttnpb.ApplicationIdentifiers][]ttnpb.Right{
			{ApplicationID: "foo"}: {ttnpb.RIGHT_APPLICATION_INFO},
		},
		GatewayRights: map[ttnpb.GatewayIdentifiers][]ttnpb.Right{
			{GatewayID: "foo"}: {ttnpb.RIGHT_GATEWAY_INFO},
		},
		OrganizationRights: map[ttnpb.OrganizationIdentifiers][]ttnpb.Right{
			{OrganizationID: "foo"}: {ttnpb.RIGHT_ORGANIZATION_INFO},
		},
	})

	fooRes := requireRights(fooCtx, "foo")
	a.So(fooRes.AppErr, should.BeNil)
	a.So(fooRes.GtwErr, should.BeNil)
	a.So(fooRes.OrgErr, should.BeNil)

	barRes := requireRights(fooCtx, "bar")
	a.So(errors.IsPermissionDenied(barRes.AppErr), should.BeTrue)
	a.So(errors.IsPermissionDenied(barRes.GtwErr), should.BeTrue)
	a.So(errors.IsPermissionDenied(barRes.OrgErr), should.BeTrue)

	mockErr := errors.New("mock")
	errFetchCtx := NewContextWithFetcher(context.Background(), &mockFetcher{
		applicationError:  mockErr,
		gatewayError:      mockErr,
		organizationError: mockErr,
	})
	errFetchRes := requireRights(errFetchCtx, "foo")
	a.So(errFetchRes.AppErr, should.Resemble, mockErr)
	a.So(errFetchRes.GtwErr, should.Resemble, mockErr)
	a.So(errFetchRes.OrgErr, should.Resemble, mockErr)

	errPermissionDenied := status.New(codes.PermissionDenied, "permission denied").Err()
	permissionDeniedFetchCtx := NewContextWithFetcher(context.Background(), &mockFetcher{
		applicationError:  errPermissionDenied,
		gatewayError:      errPermissionDenied,
		organizationError: errPermissionDenied,
	})
	permissionDeniedRes := requireRights(permissionDeniedFetchCtx, "foo")
	a.So(errors.IsPermissionDenied(permissionDeniedRes.AppErr), should.BeTrue)
	a.So(errors.IsPermissionDenied(permissionDeniedRes.GtwErr), should.BeTrue)
	a.So(errors.IsPermissionDenied(permissionDeniedRes.OrgErr), should.BeTrue)
}