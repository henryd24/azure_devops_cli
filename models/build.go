package models

import "time"

type BuildDefinition struct {
	Links                     *BuildDefinitionLinks      `json:"_links,omitempty"`
	AuthoredBy                *AuthoredBy                `json:"authoredBy,omitempty"`
	CreatedDate               *time.Time                 `json:"createdDate,omitempty"`
	Drafts                    []interface{}              `json:"drafts,omitempty"`
	ID                        *int64                     `json:"id,omitempty"`
	JobAuthorizationScope     *string                    `json:"jobAuthorizationScope,omitempty"`
	JobCancelTimeoutInMinutes *int64                     `json:"jobCancelTimeoutInMinutes,omitempty"`
	JobTimeoutInMinutes       *int64                     `json:"jobTimeoutInMinutes,omitempty"`
	Name                      *string                    `json:"name,omitempty"`
	Options                   []Option                   `json:"options,omitempty"`
	Path                      *string                    `json:"path,omitempty"`
	Process                   *Process                   `json:"process,omitempty"`
	Project                   *BuildDefinitionProject    `json:"project,omitempty"`
	Properties                *BuildDefinitionProperties `json:"properties,omitempty"`
	Quality                   *string                    `json:"quality,omitempty"`
	Queue                     *Queue                     `json:"queue,omitempty"`
	QueueStatus               *string                    `json:"queueStatus,omitempty"`
	Repository                *BuildDefinitionRepository `json:"repository,omitempty"`
	Revision                  *int64                     `json:"revision,omitempty"`
	Tags                      []interface{}              `json:"tags,omitempty"`
	Triggers                  []Trigger                  `json:"triggers,omitempty"`
	Type                      *string                    `json:"type,omitempty"`
	URI                       *string                    `json:"uri,omitempty"`
	URL                       *string                    `json:"url,omitempty"`
}

type AuthoredBy struct {
	Links       *AuthoredByLinks `json:"_links,omitempty"`
	Descriptor  *string          `json:"descriptor,omitempty"`
	DisplayName *string          `json:"displayName,omitempty"`
	ID          *string          `json:"id,omitempty"`
	ImageURL    *string          `json:"imageUrl,omitempty"`
	UniqueName  *string          `json:"uniqueName,omitempty"`
	URL         *string          `json:"url,omitempty"`
}

type AuthoredByLinks struct {
	Avatar *Badge `json:"avatar,omitempty"`
}

type Badge struct {
	Href *string `json:"href,omitempty"`
}

type BuildDefinitionLinks struct {
	Badge  *Badge `json:"badge,omitempty"`
	Editor *Badge `json:"editor,omitempty"`
	Self   *Badge `json:"self,omitempty"`
	Web    *Badge `json:"web,omitempty"`
}

type Option struct {
	Definition *Definition `json:"definition,omitempty"`
	Enabled    *bool       `json:"enabled,omitempty"`
	Inputs     *Inputs     `json:"inputs,omitempty"`
}

type Definition struct {
	ID *string `json:"id,omitempty"`
}

type Inputs struct {
	AdditionalFields  *string `json:"additionalFields,omitempty"`
	BranchFilters     *string `json:"branchFilters,omitempty"`
	AssignToRequestor *string `json:"assignToRequestor,omitempty"`
	WorkItemType      *string `json:"workItemType,omitempty"`
}

type Process struct {
	Type         *int64  `json:"type,omitempty"`
	YAMLFilename *string `json:"yamlFilename,omitempty"`
}

type BuildDefinitionProject struct {
	ID             *string    `json:"id,omitempty"`
	LastUpdateTime *time.Time `json:"lastUpdateTime,omitempty"`
	Name           *string    `json:"name,omitempty"`
	Revision       *int64     `json:"revision,omitempty"`
	State          *string    `json:"state,omitempty"`
	URL            *string    `json:"url,omitempty"`
	Visibility     *string    `json:"visibility,omitempty"`
}

type BuildDefinitionProperties struct {
}

type Queue struct {
	Links *QueueLinks `json:"_links,omitempty"`
	ID    *int64      `json:"id,omitempty"`
	Name  *string     `json:"name,omitempty"`
	Pool  *Pool       `json:"pool,omitempty"`
	URL   *string     `json:"url,omitempty"`
}

type QueueLinks struct {
	Self *Badge `json:"self,omitempty"`
}

type Pool struct {
	ID       *int64  `json:"id,omitempty"`
	IsHosted *bool   `json:"isHosted,omitempty"`
	Name     *string `json:"name,omitempty"`
}

type BuildDefinitionRepository struct {
	CheckoutSubmodules *bool                 `json:"checkoutSubmodules,omitempty"`
	Clean              *string               `json:"clean,omitempty"`
	DefaultBranch      *string               `json:"defaultBranch,omitempty"`
	ID                 *string               `json:"id,omitempty"`
	Name               *string               `json:"name,omitempty"`
	Properties         *RepositoryProperties `json:"properties,omitempty"`
	Type               *string               `json:"type,omitempty"`
	URL                *string               `json:"url,omitempty"`
}

type RepositoryProperties struct {
	APIURL                   *string `json:"apiUrl,omitempty"`
	Archived                 *string `json:"archived,omitempty"`
	BranchesURL              *string `json:"branchesUrl,omitempty"`
	CheckoutNestedSubmodules *string `json:"checkoutNestedSubmodules,omitempty"`
	CleanOptions             *string `json:"cleanOptions,omitempty"`
	CloneURL                 *string `json:"cloneUrl,omitempty"`
	ConnectedServiceID       *string `json:"connectedServiceId,omitempty"`
	DefaultBranch            *string `json:"defaultBranch,omitempty"`
	ExternalID               *string `json:"externalId,omitempty"`
	FetchDepth               *string `json:"fetchDepth,omitempty"`
	FetchTags                *string `json:"fetchTags,omitempty"`
	FullName                 *string `json:"fullName,omitempty"`
	GitLFSSupport            *string `json:"gitLfsSupport,omitempty"`
	HasAdminPermissions      *string `json:"hasAdminPermissions,omitempty"`
	IsFork                   *string `json:"isFork,omitempty"`
	IsPrivate                *string `json:"isPrivate,omitempty"`
	LabelSources             *string `json:"labelSources,omitempty"`
	LabelSourcesFormat       *string `json:"labelSourcesFormat,omitempty"`
	LastUpdated              *string `json:"lastUpdated,omitempty"`
	ManageURL                *string `json:"manageUrl,omitempty"`
	NodeID                   *string `json:"nodeId,omitempty"`
	OrgName                  *string `json:"orgName,omitempty"`
	OwnerAvatarURL           *string `json:"ownerAvatarUrl,omitempty"`
	OwnerID                  *string `json:"ownerId,omitempty"`
	OwnerIsAUser             *string `json:"ownerIsAUser,omitempty"`
	RefsURL                  *string `json:"refsUrl,omitempty"`
	ReportBuildStatus        *string `json:"reportBuildStatus,omitempty"`
	SafeRepository           *string `json:"safeRepository,omitempty"`
	ShortName                *string `json:"shortName,omitempty"`
	SkipSyncSource           *string `json:"skipSyncSource,omitempty"`
}

type Trigger struct {
	BatchChanges                                      *bool                    `json:"batchChanges,omitempty"`
	BranchFilters                                     []string                 `json:"branchFilters,omitempty"`
	MaxConcurrentBuildsPerBranch                      *int64                   `json:"maxConcurrentBuildsPerBranch,omitempty"`
	PathFilters                                       []interface{}            `json:"pathFilters,omitempty"`
	SettingsSourceType                                *int64                   `json:"settingsSourceType,omitempty"`
	TriggerType                                       *string                  `json:"triggerType,omitempty"`
	Forks                                             *Forks                   `json:"forks,omitempty"`
	IsCommentRequiredForPullRequest                   *bool                    `json:"isCommentRequiredForPullRequest,omitempty"`
	PipelineTriggerSettings                           *PipelineTriggerSettings `json:"pipelineTriggerSettings,omitempty"`
	RequireCommentsForNonTeamMemberAndNonContributors *bool                    `json:"requireCommentsForNonTeamMemberAndNonContributors,omitempty"`
	RequireCommentsForNonTeamMembersOnly              *bool                    `json:"requireCommentsForNonTeamMembersOnly,omitempty"`
}

type Forks struct {
	AllowFullAccessToken *bool `json:"allowFullAccessToken,omitempty"`
	AllowSecrets         *bool `json:"allowSecrets,omitempty"`
	Enabled              *bool `json:"enabled,omitempty"`
}

type PipelineTriggerSettings struct {
	BuildsEnabledForForks                             *bool `json:"buildsEnabledForForks,omitempty"`
	EnforceJobAuthScopeForForks                       *bool `json:"enforceJobAuthScopeForForks,omitempty"`
	EnforceNoAccessToSecretsFromForks                 *bool `json:"enforceNoAccessToSecretsFromForks,omitempty"`
	ForkProtectionEnabled                             *bool `json:"forkProtectionEnabled,omitempty"`
	IsCommentRequiredForPullRequest                   *bool `json:"isCommentRequiredForPullRequest,omitempty"`
	RequireCommentsForNonTeamMemberAndNonContributors *bool `json:"requireCommentsForNonTeamMemberAndNonContributors,omitempty"`
	RequireCommentsForNonTeamMembersOnly              *bool `json:"requireCommentsForNonTeamMembersOnly,omitempty"`
}

type BuildRunPayload struct {
	Definition struct {
		ID int `json:"id"`
	} `json:"definition,omitempty"`
	TemplateParameters map[string]string        `json:"templateParameters,omitempty"`
	Variables          map[string]BuildVariable `json:"variables,omitempty"`
}

type Build struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
	Result string `json:"result"`
	Links  struct {
		Web struct {
			Href string `json:"href"`
		} `json:"web"`
	} `json:"_links"`
}

type BuildVariable struct {
	Value    string `json:"value"`
	IsSecret bool   `json:"isSecret,omitempty"`
}
