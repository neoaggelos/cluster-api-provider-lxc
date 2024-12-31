package incus

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/lxc/incus/v6/shared/api"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// InitProfile creates an LXC profile if it does not already exist.
func (c *Client) InitProfile(ctx context.Context, profile api.ProfilesPost) error {
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("profileName", profile.Name))

	if err := c.Client.CreateProfile(profile); err != nil {
		switch {
		case strings.Contains(err.Error(), "The profile already exists"):
			log.FromContext(ctx).V(4).Info("The profile already exists")
			return nil
		case strings.Contains(err.Error(), "Privileged containers are forbidden"):
			// TODO: handle case of restricted projects, e.g.
			//
			// $ incus profile set t1 security.privileged=true
			// Error: Failed checking if profile update allowed: Invalid value "true" for config "security.privileged" on  "t1" of project "user-1000": Privileged containers are forbidden
			return terminalError{err}
		}
		return fmt.Errorf("failed to CreateProfile: %w", err)
	} else {
		log.FromContext(ctx).V(4).Info("Successfully created profile")
	}
	return nil
}

// DeleteProfile deletes an LXC profile if it exists.
func (c *Client) DeleteProfile(ctx context.Context, profileName string) error {
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("profileName", profileName))

	if err := c.Client.DeleteProfile(profileName); err != nil {
		if strings.Contains(err.Error(), "Profile not found") {
			log.FromContext(ctx).V(4).Info("The profile does not exist")
			return nil
		}

		return fmt.Errorf("failed to DeleteProfile: %w", err)
	}

	log.FromContext(ctx).V(4).Info("Successfully removed profile")
	return nil
}
