package gitlab

import (
	"fmt"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// GetAllProjects retrieves all accessible projects, optionally filtered by group
func (c *Client) GetAllProjects(groupName string) ([]*gitlab.Project, error) {
	if groupName != "" {
		return c.getGroupProjects(groupName)
	}
	return c.getAllAccessibleProjects()
}

func (c *Client) getGroupProjects(groupName string) ([]*gitlab.Project, error) {
	var allProjects []*gitlab.Project

	// First, try to get the group to verify it exists and get its ID/path
	group, _, err := c.Groups.GetGroup(groupName, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting group %s: %w (make sure the group path/name is correct)", groupName, err)
	}

	c.logger.Infof("Found group: %s (ID: %d, Path: %s)", group.Name, group.ID, group.FullPath)

	// Try different combinations to find projects
	// First try with IncludeSubGroups and without Archived filter
	groupOpt := &gitlab.ListGroupProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
			Page:    1,
		},
		IncludeSubGroups: gitlab.Ptr(true), // Include projects from subgroups
	}

	// Use group ID directly (as int, which is more reliable)
	c.logger.Infof("Fetching projects for group ID: %d (include_subgroups: true, all projects)", group.ID)

	for {
		projects, resp, err := c.Groups.ListGroupProjects(group.ID, groupOpt)
		if err != nil {
			return nil, fmt.Errorf("error getting projects for group %d (%s): %w", group.ID, group.FullPath, err)
		}

		c.logger.Infof("Page %d: found %d projects", groupOpt.Page, len(projects))
		allProjects = append(allProjects, projects...)

		if resp.NextPage == 0 {
			break
		}

		groupOpt.Page = resp.NextPage
	}

	c.logger.Infof("Total projects found in group (including subgroups): %d", len(allProjects))

	// If no projects found, try without IncludeSubGroups to see if there are projects in the group itself
	if len(allProjects) == 0 {
		c.logger.Warn("No projects found with subgroups, trying without IncludeSubGroups...")
		groupOpt.IncludeSubGroups = gitlab.Ptr(false)
		groupOpt.Page = 1

		for {
			projects, resp, err := c.Groups.ListGroupProjects(group.ID, groupOpt)
			if err != nil {
				c.logger.Warnf("Error getting projects without subgroups: %v", err)
				break
			}

			c.logger.Infof("Page %d (no subgroups): found %d projects", groupOpt.Page, len(projects))
			allProjects = append(allProjects, projects...)

			if resp.NextPage == 0 {
				break
			}

			groupOpt.Page = resp.NextPage
		}

		c.logger.Infof("Total projects found in group (without subgroups): %d", len(allProjects))
	}

	return allProjects, nil
}

func (c *Client) getAllAccessibleProjects() ([]*gitlab.Project, error) {
	var allProjects []*gitlab.Project

	// Try multiple approaches to get all projects
	c.logger.Info("Fetching all accessible projects...")

	// First, try without any filters to get all projects
	opt := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
			Page:    1,
		},
		Archived: gitlab.Ptr(true), // Include archived
		Simple:   gitlab.Ptr(false),
	}

	for {
		projects, resp, err := c.Projects.ListProjects(opt)
		if err != nil {
			return nil, fmt.Errorf("error getting project list: %w", err)
		}

		c.logger.Infof("Page %d: found %d projects (total so far: %d)", opt.Page, len(projects), len(allProjects)+len(projects))
		allProjects = append(allProjects, projects...)

		if resp.NextPage == 0 {
			c.logger.Infof("No more pages. Total projects: %d", len(allProjects))
			break
		}

		opt.Page = resp.NextPage
	}

	// If we got less than expected, try with Membership=true to get projects where user is a member
	if len(allProjects) < 200 {
		c.logger.Warnf("Only found %d projects, trying with Membership=true to get member projects...", len(allProjects))
		memberOpt := &gitlab.ListProjectsOptions{
			ListOptions: gitlab.ListOptions{
				PerPage: 100,
				Page:    1,
			},
			Archived:   gitlab.Ptr(true),
			Simple:     gitlab.Ptr(false),
			Membership: gitlab.Ptr(true),
		}

		memberProjects := make(map[int]*gitlab.Project)
		for _, p := range allProjects {
			memberProjects[p.ID] = p
		}

		for {
			projects, resp, err := c.Projects.ListProjects(memberOpt)
			if err != nil {
				c.logger.Warnf("Error getting member projects: %v", err)
				break
			}

			c.logger.Infof("Membership page %d: found %d projects", memberOpt.Page, len(projects))
			for _, p := range projects {
				if _, exists := memberProjects[p.ID]; !exists {
					allProjects = append(allProjects, p)
					memberProjects[p.ID] = p
				}
			}

			if resp.NextPage == 0 {
				break
			}

			memberOpt.Page = resp.NextPage
		}

		c.logger.Infof("After adding member projects: %d total projects", len(allProjects))
	}

	return allProjects, nil
}
