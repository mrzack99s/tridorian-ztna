package services

import (
	"context"
	"fmt"
	"tridorian-ztna/internal/models"

	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

type IdentityService struct{}

func NewIdentityService() *IdentityService {
	return &IdentityService{}
}

// FetchGoogleUsers pulls users from a Google Workspace domain.
func (s *IdentityService) FetchGoogleUsers(ctx context.Context, serviceAccountJSON []byte, adminEmail string, domain string) ([]models.ExternalIdentity, error) {
	config, err := google.JWTConfigFromJSON(serviceAccountJSON, admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service account JSON: %w", err)
	}
	config.Subject = adminEmail

	ts := config.TokenSource(ctx)
	srv, err := admin.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("failed to create admin service: %w", err)
	}

	var allUsers []models.ExternalIdentity
	pageToken := ""

	for {
		call := srv.Users.List().Context(ctx).MaxResults(500)
		if domain != "" {
			call = call.Domain(domain)
		} else {
			call = call.Customer("my_customer")
		}

		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("failed to list users: %w", err)
		}

		for _, u := range resp.Users {
			allUsers = append(allUsers, models.ExternalIdentity{
				Name:       u.Name.FullName,
				Email:      u.PrimaryEmail,
				ExternalID: u.Id,
				IsAdmin:    u.IsAdmin,
			})
		}

		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return allUsers, nil
}

// SearchGoogleUsers performs a search for users, ideal for Auto Complete.
func (s *IdentityService) SearchGoogleUsers(ctx context.Context, serviceAccountJSON []byte, adminEmail string, queryStr string) ([]models.ExternalIdentity, error) {
	config, err := google.JWTConfigFromJSON(serviceAccountJSON, admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service account JSON: %w", err)
	}
	config.Subject = adminEmail

	ts := config.TokenSource(ctx)
	srv, err := admin.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("failed to create admin service: %w", err)
	}

	googleQuery := fmt.Sprintf("name:'%s'* OR email:'%s'*", queryStr, queryStr)

	resp, err := srv.Users.List().
		Context(ctx).
		Customer("my_customer").
		MaxResults(20).
		Query(googleQuery).
		OrderBy("email").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	var foundUsers []models.ExternalIdentity
	for _, u := range resp.Users {
		foundUsers = append(foundUsers, models.ExternalIdentity{
			Name:       u.Name.FullName,
			Email:      u.PrimaryEmail,
			ExternalID: u.Id,
			IsAdmin:    u.IsAdmin,
		})
	}

	return foundUsers, nil
}

// GetGoogleAdmins fetches only users who have administrator privileges in Google Workspace.
func (s *IdentityService) GetGoogleAdmins(ctx context.Context, serviceAccountJSON []byte, adminEmail string) ([]models.ExternalIdentity, error) {
	config, err := google.JWTConfigFromJSON(serviceAccountJSON, admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service account JSON: %w", err)
	}
	config.Subject = adminEmail

	ts := config.TokenSource(ctx)
	srv, err := admin.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("failed to create admin service: %w", err)
	}

	resp, err := srv.Users.List().
		Context(ctx).
		Customer("my_customer").
		Query("isAdmin=true").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list admins: %w", err)
	}

	var admins []models.ExternalIdentity
	for _, u := range resp.Users {
		admins = append(admins, models.ExternalIdentity{
			Name:       u.Name.FullName,
			Email:      u.PrimaryEmail,
			ExternalID: u.Id,
			IsAdmin:    true,
		})
	}

	return admins, nil
}

func (s *IdentityService) GetUserGroups(ctx context.Context, serviceAccountJSON []byte, adminEmail, userEmail string) ([]string, error) {
	// We need both User and Group scopes to check group memberships
	config, err := google.JWTConfigFromJSON(serviceAccountJSON,
		admin.AdminDirectoryUserReadonlyScope,
		admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service account JSON: %w", err)
	}
	config.Subject = adminEmail

	ts := config.TokenSource(ctx)
	srv, err := admin.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("failed to create admin service: %w", err)
	}

	resp, err := srv.Groups.List().
		Context(ctx).
		UserKey(userEmail).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list user groups: %w", err)
	}

	var groups []string
	for _, g := range resp.Groups {
		groups = append(groups, g.Email)
	}

	return groups, nil
}
