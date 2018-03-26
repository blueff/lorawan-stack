// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"context"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/email/mock"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/email/templates"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var _ ttnpb.IsAdminServer = new(adminService)

func TestAdminSettings(t *testing.T) {
	a := assertions.New(t)
	is := getIS(t)
	defer is.store.Settings.Set(testSettings())

	ctx := testCtx(testUsers()["alice"].UserID)

	resp, err := is.adminService.GetSettings(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	a.So(resp, test.ShouldBeSettingsIgnoringAutoFields, testSettings())

	// modify settings
	_, err = is.adminService.UpdateSettings(ctx, &ttnpb.UpdateSettingsRequest{
		Settings: ttnpb.IdentityServerSettings{
			IdentityServerSettings_UserRegistrationFlow: ttnpb.IdentityServerSettings_UserRegistrationFlow{
				SkipValidation: true,
			},
		},
		UpdateMask: pbtypes.FieldMask{
			Paths: []string{"allowed_emails", "user_registration.skip_validation"},
		},
	})
	a.So(err, should.BeNil)

	resp, err = is.GetSettings(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	a.So(resp.AllowedEmails, should.HaveLength, 0)
	a.So(resp.IdentityServerSettings_UserRegistrationFlow.SkipValidation, should.BeTrue)
}

func TestAdminInvitations(t *testing.T) {
	a := assertions.New(t)
	is := getIS(t)

	ctx := testCtx(testUsers()["alice"].UserID)
	email := "bar@baz.com"

	_, err := is.adminService.SendInvitation(ctx, &ttnpb.SendInvitationRequest{Email: email})
	a.So(err, should.BeNil)

	// gather the token to register an account
	token := ""

	invitation, ok := mock.Data().(*templates.Invitation)
	if a.So(ok, should.BeTrue) {
		token = invitation.Token
		a.So(invitation.Token, should.NotBeEmpty)
	}

	invitations, err := is.adminService.ListInvitations(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	if a.So(invitations.Invitations, should.HaveLength, 1) {
		i := invitations.Invitations[0]
		a.So(i.Email, should.Equal, email)
		a.So(i.IssuedAt.IsZero(), should.BeFalse)
		a.So(i.ExpiresAt.IsZero(), should.BeFalse)
	}

	// use invitation
	settings, err := is.store.Settings.Get()
	a.So(err, should.BeNil)
	defer func() {
		settings.IdentityServerSettings_UserRegistrationFlow.InvitationOnly = false
		is.store.Settings.Set(settings)
	}()

	settings.IdentityServerSettings_UserRegistrationFlow.InvitationOnly = true
	err = is.store.Settings.Set(settings)
	a.So(err, should.BeNil)

	user := ttnpb.User{
		UserIdentifier: ttnpb.UserIdentifier{UserID: "invitation-user"},
		Password:       "lol",
		Email:          email,
		Name:           "HI",
	}

	_, err = is.userService.CreateUser(context.Background(), &ttnpb.CreateUserRequest{User: user})
	a.So(err, should.NotBeNil)
	a.So(ErrInvitationTokenMissing.Describes(err), should.BeTrue)

	_, err = is.userService.CreateUser(context.Background(), &ttnpb.CreateUserRequest{
		User:            user,
		InvitationToken: token,
	})
	a.So(err, should.BeNil)
	defer is.store.Users.Delete(user.UserID)

	// check user was created
	found, err := is.adminService.GetUser(ctx, &ttnpb.UserIdentifier{UserID: user.UserID})
	a.So(err, should.BeNil)
	a.So(found.UserID, should.Equal, user.UserID)
	a.So(found.Password, should.BeEmpty)

	// check invitation was used
	invitations, err = is.adminService.ListInvitations(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	a.So(invitations.Invitations, should.HaveLength, 0)

	// can't send invitation to an already used email address
	_, err = is.adminService.SendInvitation(ctx, &ttnpb.SendInvitationRequest{Email: email})
	a.So(err, should.NotBeNil)
	a.So(ErrEmailAddressAlreadyUsed.Describes(err), should.BeTrue)

	// send a new invitation but revoke it later
	email = "bar@bazqux.com"
	_, err = is.adminService.SendInvitation(ctx, &ttnpb.SendInvitationRequest{Email: email})
	a.So(err, should.BeNil)

	invitations, err = is.adminService.ListInvitations(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	if a.So(invitations.Invitations, should.HaveLength, 1) {
		i := invitations.Invitations[0]
		a.So(i.Email, should.Equal, email)
		a.So(i.IssuedAt.IsZero(), should.BeFalse)
		a.So(i.ExpiresAt.IsZero(), should.BeFalse)
	}

	_, err = is.adminService.DeleteInvitation(ctx, &ttnpb.DeleteInvitationRequest{Email: email})
	a.So(err, should.BeNil)

	invitations, err = is.adminService.ListInvitations(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	a.So(invitations.Invitations, should.HaveLength, 0)
}

func TestAdminUsers(t *testing.T) {
	a := assertions.New(t)
	is := getIS(t)

	ctx := testCtx(testUsers()["alice"].UserID)
	user := testUsers()["bob"]

	// reset password
	found, err := is.store.Users.GetByID(user.UserID, is.config.Specializers.User)
	a.So(err, should.BeNil)

	old := found.GetUser().Password

	{
		resp, err := is.adminService.ResetUserPassword(ctx, &ttnpb.UserIdentifier{UserID: user.UserID})
		a.So(err, should.BeNil)
		a.So(resp.Password, should.NotBeEmpty)

		data, ok := mock.Data().(*templates.PasswordReset)
		if a.So(ok, should.BeTrue) {
			a.So(data.Password, should.NotBeEmpty)
			a.So(data.Password, should.Equal, resp.Password)
		}

		found, err = is.store.Users.GetByID(user.UserID, is.config.Specializers.User)
		a.So(err, should.BeNil)
		a.So(old, should.NotEqual, found.GetUser().Password)
	}

	// make user admin
	_, err = is.adminService.UpdateUser(ctx, &ttnpb.UpdateUserRequest{
		User: ttnpb.User{
			UserIdentifier: ttnpb.UserIdentifier{UserID: user.UserID},
			Admin:          true,
		},
		UpdateMask: pbtypes.FieldMask{
			Paths: []string{"admin"},
		},
	})
	a.So(err, should.BeNil)

	found, err = is.store.Users.GetByID(user.UserID, is.config.Specializers.User)
	a.So(err, should.BeNil)
	a.So(found.GetUser().Admin, should.BeTrue)

	// delete user
	user.UserID = "tmp-user"
	user.Email = "fofofo@bar.com"
	err = is.store.Users.Create(user)
	a.So(err, should.BeNil)

	_, err = is.adminService.DeleteUser(ctx, &ttnpb.UserIdentifier{UserID: user.UserID})
	a.So(err, should.BeNil)

	ddata, ok := mock.Data().(*templates.AccountDeleted)
	if a.So(ok, should.BeTrue) {
		a.So(ddata.UserID, should.Equal, user.UserID)
	}

	_, err = is.store.Users.GetByID(user.UserID, is.config.Specializers.User)
	a.So(err, should.NotBeNil)
	a.So(sql.ErrUserNotFound.Describes(err), should.BeTrue)

	// list approved users
	resp, err := is.adminService.ListUsers(ctx, &ttnpb.ListUsersRequest{
		ListUsersRequest_FilterState: &ttnpb.ListUsersRequest_FilterState{State: ttnpb.STATE_APPROVED},
	})
	a.So(err, should.BeNil)
	if a.So(resp.Users, should.HaveLength, 1) {
		a.So(resp.Users[0], test.ShouldBeUserIgnoringAutoFields, testUsers()["alice"])
	}
}

func TestAdminClients(t *testing.T) {
	a := assertions.New(t)
	is := getIS(t)

	ctx := testCtx(testUsers()["alice"].UserID)
	client := testClient()

	found, err := is.adminService.GetClient(ctx, &ttnpb.ClientIdentifier{ClientID: client.ClientID})
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeClientIgnoringAutoFields, client)

	clients, err := is.adminService.ListClients(ctx, &ttnpb.ListClientsRequest{
		ListClientsRequest_FilterState: &ttnpb.ListClientsRequest_FilterState{State: ttnpb.STATE_PENDING},
	})
	a.So(err, should.BeNil)
	a.So(clients.Clients, should.HaveLength, 0)

	client.ClientID = "tmp-client"

	err = is.store.Clients.Create(client)
	a.So(err, should.BeNil)

	_, err = is.adminService.DeleteClient(ctx, &ttnpb.ClientIdentifier{ClientID: client.ClientID})
	a.So(err, should.BeNil)

	data, ok := mock.Data().(*templates.ClientDeleted)
	if a.So(ok, should.BeTrue) {
		a.So(data.ClientID, should.Equal, "tmp-client")
	}
}
