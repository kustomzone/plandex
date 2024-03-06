package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"plandex-server/db"
	"plandex-server/types"
	"strings"

	"github.com/gorilla/mux"
	"github.com/plandex/plandex/shared"
)

func InviteUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received a request for InviteUserHandler")
	auth := Authenticate(w, r, true)
	if auth == nil {
		return
	}

	if auth.User.IsTrial {
		WriteApiError(w, shared.ApiError{
			Type:   shared.ApiErrorTypeTrialActionNotAllowed,
			Status: http.StatusForbidden,
			Msg:    "Anonymous trial user can't invite other users",
		})

		return
	}

	currentUserId := auth.User.Id

	var req shared.InviteRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("Error unmarshalling request: %v\n", err)
		http.Error(w, "Error unmarshalling request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	req.Email = strings.ToLower(req.Email)

	// ensure current user can invite target user
	permission := types.Permission(strings.Join([]string{string(types.PermissionInviteUser), req.OrgRoleId}, "|"))

	if !auth.HasPermission(permission) {
		log.Printf("User does not have permission to invite user with role: %v\n", req.OrgRoleId)
		http.Error(w, "User does not have permission to invite user with role: "+req.OrgRoleId, http.StatusForbidden)
		return
	}

	// ensure user doesn't already have access to org via domain
	split := strings.Split(req.Email, "@")
	if len(split) != 2 {
		log.Printf("Invalid email: %v\n", req.Email)
		http.Error(w, "Invalid email: "+req.Email, http.StatusBadRequest)
		return
	}
	domain := &split[1]
	org, err := db.GetOrg(auth.OrgId)

	if err != nil {
		log.Printf("Error getting org: %v\n", err)
		http.Error(w, "Error getting org: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if org.AutoAddDomainUsers && org.Domain == domain {
		log.Printf("User already has access to org via domain: %v\n", domain)
		http.Error(w, "User already has access to org via domain: "+*domain, http.StatusBadRequest)
	}

	// ensure user with this email isn't already in the org
	user, err := db.GetUserByEmail(req.Email)

	if err != nil {
		log.Printf("Error getting user: %v\n", err)
		http.Error(w, "Error getting user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if user != nil {
		isMember, err := db.ValidateOrgMembership(user.Id, auth.OrgId)

		if err != nil {
			log.Printf("Error validating org membership: %v\n", err)
			http.Error(w, "Error validating org membership: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if isMember {
			log.Println("User is already a member of org")
			http.Error(w, "User is already a member of org", http.StatusBadRequest)
			return
		}
	}

	// ensure invite isn't already active
	invite, err := db.GetInviteForOrgUser(auth.OrgId, req.Email)

	if err != nil {
		log.Printf("Error getting invite: %v\n", err)
		http.Error(w, "Error getting invite: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if invite != nil {
		log.Println("Invite already exists")
		http.Error(w, "Invite already exists", http.StatusBadRequest)
		return
	}

	err = db.CreateInvite(&db.Invite{
		OrgId:     auth.OrgId,
		OrgRoleId: req.OrgRoleId,
		Email:     req.Email,
		Name:      req.Name,
		InviterId: currentUserId,
	})

	if err != nil {
		log.Printf("Error creating invite: %v\n", err)
		http.Error(w, "Error creating invite: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Successfully created invite")
}

func ListPendingInvitesHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received a request for ListInvitesHandler")
	auth := Authenticate(w, r, true)
	if auth == nil {
		return
	}

	if auth.User.IsTrial {
		WriteApiError(w, shared.ApiError{
			Type:   shared.ApiErrorTypeTrialActionNotAllowed,
			Status: http.StatusForbidden,
			Msg:    "Anonymous trial user can't list invites",
		})
		return
	}

	invites, err := db.ListPendingInvites(auth.OrgId)

	if err != nil {
		log.Printf("Error listing invites: %v\n", err)
		http.Error(w, "Error listing invites: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var apiInvites []*shared.Invite
	for _, invite := range invites {
		apiInvites = append(apiInvites, invite.ToApi())
	}

	bytes, err := json.Marshal(apiInvites)

	if err != nil {
		log.Printf("Error marshalling invites: %v\n", err)
		http.Error(w, "Error marshalling invites: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
	log.Println("Successfully processed request for ListPendingInvitesHandler")
}

func ListAcceptedInvitesHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received a request for ListAcceptedInvitesHandler")
	auth := Authenticate(w, r, true)
	if auth == nil {
		return
	}

	if auth.User.IsTrial {
		WriteApiError(w, shared.ApiError{
			Type:   shared.ApiErrorTypeTrialActionNotAllowed,
			Status: http.StatusForbidden,
			Msg:    "Anonymous trial user can't list invites",
		})
		return
	}

	invites, err := db.ListAcceptedInvites(auth.OrgId)

	if err != nil {
		log.Printf("Error listing invites: %v\n", err)
		http.Error(w, "Error listing invites: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var apiInvites []*shared.Invite
	for _, invite := range invites {
		apiInvites = append(apiInvites, invite.ToApi())
	}

	bytes, err := json.Marshal(apiInvites)

	if err != nil {
		log.Printf("Error marshalling invites: %v\n", err)
		http.Error(w, "Error marshalling invites: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
	log.Println("Successfully processed request for ListAcceptedInvitesHandler")
}

func ListAllInvitesHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received a request for ListAllInvitesHandler")
	auth := Authenticate(w, r, true)
	if auth == nil {
		return
	}

	if auth.User.IsTrial {
		WriteApiError(w, shared.ApiError{
			Type:   shared.ApiErrorTypeTrialActionNotAllowed,
			Status: http.StatusForbidden,
			Msg:    "Anonymous trial user can't list invites",
		})
		return
	}

	invites, err := db.ListAllInvites(auth.OrgId)

	if err != nil {
		log.Printf("Error listing invites: %v\n", err)
		http.Error(w, "Error listing invites: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var apiInvites []*shared.Invite
	for _, invite := range invites {
		apiInvites = append(apiInvites, invite.ToApi())
	}

	bytes, err := json.Marshal(apiInvites)

	if err != nil {
		log.Printf("Error marshalling invites: %v\n", err)
		http.Error(w, "Error marshalling invites: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
	log.Println("Successfully processed request for ListAllInvitesHandler")
}

func DeleteInviteHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received a request for DeleteInviteHandler")
	auth := Authenticate(w, r, true)
	if auth == nil {
		return
	}

	if auth.User.IsTrial {
		WriteApiError(w, shared.ApiError{
			Type:   shared.ApiErrorTypeTrialActionNotAllowed,
			Status: http.StatusForbidden,
			Msg:    "Anonymous trial user can't delete invites",
		})
		return
	}

	vars := mux.Vars(r)
	inviteId := vars["inviteId"]

	invite, err := db.GetInvite(inviteId)

	if err != nil {
		log.Printf("Error getting invite: %v\n", err)
		http.Error(w, "Error getting invite: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if invite.OrgId != auth.OrgId {
		log.Printf("Invite does not belong to org: %v\n", inviteId)
		http.Error(w, "Invite does not belong to org: "+inviteId, http.StatusBadRequest)
		return
	}

	// ensure current user can remove target invite
	removePermission := types.Permission(strings.Join([]string{string(types.PermissionRemoveUser), invite.OrgRoleId}, "|"))

	invitePermission := types.Permission(strings.Join([]string{string(types.PermissionInviteUser), invite.OrgRoleId}, "|"))

	if !(auth.HasPermission(removePermission) ||
		(auth.User.Id == invite.InviterId && auth.HasPermission(invitePermission))) {
		log.Printf("User does not have permission to remove invite with role: %v\n", invite.OrgRoleId)
		http.Error(w, "User does not have permission to remove invite with role: "+invite.OrgRoleId, http.StatusForbidden)
		return
	}

	err = db.DeleteInvite(inviteId, nil)

	if err != nil {
		log.Printf("Error deleting invite: %v\n", err)
		http.Error(w, "Error deleting invite: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Successfully deleted invite")
}
