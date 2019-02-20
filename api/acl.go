package api

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"time"
)

const (
	// ACLClientType is the client type token
	ACLClientType = "client"

	// ACLManagementType is the management type token
	ACLManagementType = "management"
)

type ACLTokenPolicyLink struct {
	ID   string
	Name string
}
type ACLTokenRoleLink struct {
	ID        string
	Name      string
	BoundName string `json:",omitempty"`
}

// ACLToken represents an ACL Token
type ACLToken struct {
	CreateIndex       uint64
	ModifyIndex       uint64
	AccessorID        string
	SecretID          string
	Description       string
	Policies          []*ACLTokenPolicyLink `json:",omitempty"`
	Roles             []*ACLTokenRoleLink   `json:",omitempty"`
	ServiceIdentities []*ACLServiceIdentity `json:",omitempty"`
	Local             bool
	ExpirationTTL     time.Duration `json:",omitempty"`
	ExpirationTime    time.Time     `json:",omitempty"`
	CreateTime        time.Time     `json:",omitempty"`
	Hash              []byte        `json:",omitempty"`

	// DEPRECATED (ACL-Legacy-Compat)
	// Rules will only be present for legacy tokens returned via the new APIs
	Rules string `json:",omitempty"`
}

type ACLTokenListEntry struct {
	CreateIndex       uint64
	ModifyIndex       uint64
	AccessorID        string
	Description       string
	Policies          []*ACLTokenPolicyLink `json:",omitempty"`
	Roles             []*ACLTokenRoleLink   `json:",omitempty"`
	ServiceIdentities []*ACLServiceIdentity `json:",omitempty"`
	Local             bool
	ExpirationTime    time.Time `json:",omitempty"`
	CreateTime        time.Time
	Hash              []byte
	Legacy            bool
}

// ACLEntry is used to represent a legacy ACL token
// The legacy tokens are deprecated.
type ACLEntry struct {
	CreateIndex uint64
	ModifyIndex uint64
	ID          string
	Name        string
	Type        string
	Rules       string
}

// ACLReplicationStatus is used to represent the status of ACL replication.
type ACLReplicationStatus struct {
	Enabled              bool
	Running              bool
	SourceDatacenter     string
	ReplicationType      string
	ReplicatedIndex      uint64
	ReplicatedRoleIndex  uint64
	ReplicatedTokenIndex uint64
	LastSuccess          time.Time
	LastError            time.Time
}

// ACLServiceIdentity represents a high-level grant of all necessary privileges
// to assume the identity of the named Service in the Catalog and within
// Connect.
type ACLServiceIdentity struct {
	ServiceName string
	Datacenters []string `json:",omitempty"`
}

// ACLPolicy represents an ACL Policy.
type ACLPolicy struct {
	ID          string
	Name        string
	Description string
	Rules       string
	Datacenters []string
	Hash        []byte
	CreateIndex uint64
	ModifyIndex uint64
}

type ACLPolicyListEntry struct {
	ID          string
	Name        string
	Description string
	Datacenters []string
	Hash        []byte
	CreateIndex uint64
	ModifyIndex uint64
}

type ACLRolePolicyLink struct {
	ID   string
	Name string
}

// ACLRole represents an ACL Role.
type ACLRole struct {
	ID                string
	Name              string
	Description       string
	Policies          []*ACLRolePolicyLink  `json:",omitempty"`
	ServiceIdentities []*ACLServiceIdentity `json:",omitempty"`
	Hash              []byte
	CreateIndex       uint64
	ModifyIndex       uint64
}

type ACLRoleBindingRule struct {
	ID          string
	Description string
	IDPName     string
	Match       []*ACLRoleBindingRuleMatch
	RoleName    string
	MustExist   bool `json:",omitempty"`

	CreateIndex uint64
	ModifyIndex uint64
}

type ACLRoleBindingRuleMatch struct {
	Selector []string
}

type ACLIdentityProvider struct {
	Name        string
	Description string
	Type        string

	KubernetesHost              string `json:",omitempty"`
	KubernetesCACert            string `json:",omitempty"`
	KubernetesServiceAccountJWT string `json:",omitempty"`

	CreateIndex uint64
	ModifyIndex uint64
}

type ACLIdentityProviderListEntry struct {
	Name        string
	Description string
	Type        string

	KubernetesHost string `json:",omitempty"`

	CreateIndex uint64
	ModifyIndex uint64
}

type ACLLoginParams struct {
	IDPType  string
	IDPName  string
	IDPToken string
	Meta     map[string]string `json:",omitempty"`
}

// ACL can be used to query the ACL endpoints
type ACL struct {
	c *Client
}

// ACL returns a handle to the ACL endpoints
func (c *Client) ACL() *ACL {
	return &ACL{c}
}

// Bootstrap is used to perform a one-time ACL bootstrap operation on a cluster
// to get the first management token.
func (a *ACL) Bootstrap() (*ACLToken, *WriteMeta, error) {
	r := a.c.newRequest("PUT", "/v1/acl/bootstrap")
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	wm := &WriteMeta{RequestTime: rtt}
	var out ACLToken
	if err := decodeBody(resp, &out); err != nil {
		return nil, nil, err
	}
	return &out, wm, nil
}

// Create is used to generate a new token with the given parameters
//
// Deprecated: Use TokenCreate instead.
func (a *ACL) Create(acl *ACLEntry, q *WriteOptions) (string, *WriteMeta, error) {
	r := a.c.newRequest("PUT", "/v1/acl/create")
	r.setWriteOptions(q)
	r.obj = acl
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	wm := &WriteMeta{RequestTime: rtt}
	var out struct{ ID string }
	if err := decodeBody(resp, &out); err != nil {
		return "", nil, err
	}
	return out.ID, wm, nil
}

// Update is used to update the rules of an existing token
//
// Deprecated: Use TokenUpdate instead.
func (a *ACL) Update(acl *ACLEntry, q *WriteOptions) (*WriteMeta, error) {
	r := a.c.newRequest("PUT", "/v1/acl/update")
	r.setWriteOptions(q)
	r.obj = acl
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	wm := &WriteMeta{RequestTime: rtt}
	return wm, nil
}

// Destroy is used to destroy a given ACL token ID
//
// Deprecated: Use TokenDelete instead.
func (a *ACL) Destroy(id string, q *WriteOptions) (*WriteMeta, error) {
	r := a.c.newRequest("PUT", "/v1/acl/destroy/"+id)
	r.setWriteOptions(q)
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	wm := &WriteMeta{RequestTime: rtt}
	return wm, nil
}

// Clone is used to return a new token cloned from an existing one
//
// Deprecated: Use TokenClone instead.
func (a *ACL) Clone(id string, q *WriteOptions) (string, *WriteMeta, error) {
	r := a.c.newRequest("PUT", "/v1/acl/clone/"+id)
	r.setWriteOptions(q)
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	wm := &WriteMeta{RequestTime: rtt}
	var out struct{ ID string }
	if err := decodeBody(resp, &out); err != nil {
		return "", nil, err
	}
	return out.ID, wm, nil
}

// Info is used to query for information about an ACL token
//
// Deprecated: Use TokenRead instead.
func (a *ACL) Info(id string, q *QueryOptions) (*ACLEntry, *QueryMeta, error) {
	r := a.c.newRequest("GET", "/v1/acl/info/"+id)
	r.setQueryOptions(q)
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	qm := &QueryMeta{}
	parseQueryMeta(resp, qm)
	qm.RequestTime = rtt

	var entries []*ACLEntry
	if err := decodeBody(resp, &entries); err != nil {
		return nil, nil, err
	}
	if len(entries) > 0 {
		return entries[0], qm, nil
	}
	return nil, qm, nil
}

// List is used to get all the ACL tokens
//
// Deprecated: Use TokenList instead.
func (a *ACL) List(q *QueryOptions) ([]*ACLEntry, *QueryMeta, error) {
	r := a.c.newRequest("GET", "/v1/acl/list")
	r.setQueryOptions(q)
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	qm := &QueryMeta{}
	parseQueryMeta(resp, qm)
	qm.RequestTime = rtt

	var entries []*ACLEntry
	if err := decodeBody(resp, &entries); err != nil {
		return nil, nil, err
	}
	return entries, qm, nil
}

// Replication returns the status of the ACL replication process in the datacenter
func (a *ACL) Replication(q *QueryOptions) (*ACLReplicationStatus, *QueryMeta, error) {
	r := a.c.newRequest("GET", "/v1/acl/replication")
	r.setQueryOptions(q)
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	qm := &QueryMeta{}
	parseQueryMeta(resp, qm)
	qm.RequestTime = rtt

	var entries *ACLReplicationStatus
	if err := decodeBody(resp, &entries); err != nil {
		return nil, nil, err
	}
	return entries, qm, nil
}

// TokenCreate creates a new ACL token. It requires that the AccessorID and SecretID fields
// of the ACLToken structure to be empty as these will be filled in by Consul.
func (a *ACL) TokenCreate(token *ACLToken, q *WriteOptions) (*ACLToken, *WriteMeta, error) {
	if token.AccessorID != "" {
		return nil, nil, fmt.Errorf("Cannot specify an AccessorID in Token Creation")
	}

	if token.SecretID != "" {
		return nil, nil, fmt.Errorf("Cannot specify a SecretID in Token Creation")
	}

	r := a.c.newRequest("PUT", "/v1/acl/token")
	r.setWriteOptions(q)
	r.obj = token
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	wm := &WriteMeta{RequestTime: rtt}
	var out ACLToken
	if err := decodeBody(resp, &out); err != nil {
		return nil, nil, err
	}

	return &out, wm, nil
}

// TokenUpdate updates a token in place without modifying its AccessorID or SecretID. A valid
// AccessorID must be set in the ACLToken structure passed to this function but the SecretID may
// be omitted and will be filled in by Consul with its existing value.
func (a *ACL) TokenUpdate(token *ACLToken, q *WriteOptions) (*ACLToken, *WriteMeta, error) {
	if token.AccessorID == "" {
		return nil, nil, fmt.Errorf("Must specify an AccessorID for Token Updating")
	}
	r := a.c.newRequest("PUT", "/v1/acl/token/"+token.AccessorID)
	r.setWriteOptions(q)
	r.obj = token
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	wm := &WriteMeta{RequestTime: rtt}
	var out ACLToken
	if err := decodeBody(resp, &out); err != nil {
		return nil, nil, err
	}

	return &out, wm, nil
}

// TokenClone will create a new token with the same policies and locality as the original
// token but will have its own auto-generated AccessorID and SecretID as well having the
// description passed to this function. The tokenID parameter must be a valid Accessor ID
// of an existing token.
func (a *ACL) TokenClone(tokenID string, description string, q *WriteOptions) (*ACLToken, *WriteMeta, error) {
	if tokenID == "" {
		return nil, nil, fmt.Errorf("Must specify a tokenID for Token Cloning")
	}

	r := a.c.newRequest("PUT", "/v1/acl/token/"+tokenID+"/clone")
	r.setWriteOptions(q)
	r.obj = struct{ Description string }{description}
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	wm := &WriteMeta{RequestTime: rtt}
	var out ACLToken
	if err := decodeBody(resp, &out); err != nil {
		return nil, nil, err
	}

	return &out, wm, nil
}

// TokenDelete removes a single ACL token. The tokenID parameter must be a valid
// Accessor ID of an existing token.
func (a *ACL) TokenDelete(tokenID string, q *WriteOptions) (*WriteMeta, error) {
	r := a.c.newRequest("DELETE", "/v1/acl/token/"+tokenID)
	r.setWriteOptions(q)
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	wm := &WriteMeta{RequestTime: rtt}
	return wm, nil
}

// TokenRead retrieves the full token details. The tokenID parameter must be a valid
// Accessor ID of an existing token.
func (a *ACL) TokenRead(tokenID string, q *QueryOptions) (*ACLToken, *QueryMeta, error) {
	r := a.c.newRequest("GET", "/v1/acl/token/"+tokenID)
	r.setQueryOptions(q)
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	qm := &QueryMeta{}
	parseQueryMeta(resp, qm)
	qm.RequestTime = rtt

	var out ACLToken
	if err := decodeBody(resp, &out); err != nil {
		return nil, nil, err
	}

	return &out, qm, nil
}

// TokenReadSelf retrieves the full token details of the token currently
// assigned to the API Client. In this manner its possible to read a token
// by its Secret ID.
func (a *ACL) TokenReadSelf(q *QueryOptions) (*ACLToken, *QueryMeta, error) {
	r := a.c.newRequest("GET", "/v1/acl/token/self")
	r.setQueryOptions(q)
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	qm := &QueryMeta{}
	parseQueryMeta(resp, qm)
	qm.RequestTime = rtt

	var out ACLToken
	if err := decodeBody(resp, &out); err != nil {
		return nil, nil, err
	}

	return &out, qm, nil
}

// TokenList lists all tokens. The listing does not contain any SecretIDs as those
// may only be retrieved by a call to TokenRead.
func (a *ACL) TokenList(q *QueryOptions) ([]*ACLTokenListEntry, *QueryMeta, error) {
	r := a.c.newRequest("GET", "/v1/acl/tokens")
	r.setQueryOptions(q)
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	qm := &QueryMeta{}
	parseQueryMeta(resp, qm)
	qm.RequestTime = rtt

	var entries []*ACLTokenListEntry
	if err := decodeBody(resp, &entries); err != nil {
		return nil, nil, err
	}
	return entries, qm, nil
}

// PolicyCreate will create a new policy. It is not allowed for the policy parameters
// ID field to be set as this will be generated by Consul while processing the request.
func (a *ACL) PolicyCreate(policy *ACLPolicy, q *WriteOptions) (*ACLPolicy, *WriteMeta, error) {
	if policy.ID != "" {
		return nil, nil, fmt.Errorf("Cannot specify an ID in Policy Creation")
	}

	r := a.c.newRequest("PUT", "/v1/acl/policy")
	r.setWriteOptions(q)
	r.obj = policy
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	wm := &WriteMeta{RequestTime: rtt}
	var out ACLPolicy
	if err := decodeBody(resp, &out); err != nil {
		return nil, nil, err
	}

	return &out, wm, nil
}

// PolicyUpdate updates a policy. The ID field of the policy parameter must be set to an
// existing policy ID
func (a *ACL) PolicyUpdate(policy *ACLPolicy, q *WriteOptions) (*ACLPolicy, *WriteMeta, error) {
	if policy.ID == "" {
		return nil, nil, fmt.Errorf("Must specify an ID in Policy Update")
	}

	r := a.c.newRequest("PUT", "/v1/acl/policy/"+policy.ID)
	r.setWriteOptions(q)
	r.obj = policy
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	wm := &WriteMeta{RequestTime: rtt}
	var out ACLPolicy
	if err := decodeBody(resp, &out); err != nil {
		return nil, nil, err
	}

	return &out, wm, nil
}

// PolicyDelete deletes a policy given its ID.
func (a *ACL) PolicyDelete(policyID string, q *WriteOptions) (*WriteMeta, error) {
	r := a.c.newRequest("DELETE", "/v1/acl/policy/"+policyID)
	r.setWriteOptions(q)
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	wm := &WriteMeta{RequestTime: rtt}
	return wm, nil
}

// PolicyRead retrieves the policy details including the rule set.
func (a *ACL) PolicyRead(policyID string, q *QueryOptions) (*ACLPolicy, *QueryMeta, error) {
	r := a.c.newRequest("GET", "/v1/acl/policy/"+policyID)
	r.setQueryOptions(q)
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	qm := &QueryMeta{}
	parseQueryMeta(resp, qm)
	qm.RequestTime = rtt

	var out ACLPolicy
	if err := decodeBody(resp, &out); err != nil {
		return nil, nil, err
	}

	return &out, qm, nil
}

// PolicyList retrieves a listing of all policies. The listing does not include the
// rules for any policy as those should be retrieved by subsequent calls to PolicyRead.
func (a *ACL) PolicyList(q *QueryOptions) ([]*ACLPolicyListEntry, *QueryMeta, error) {
	r := a.c.newRequest("GET", "/v1/acl/policies")
	r.setQueryOptions(q)
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	qm := &QueryMeta{}
	parseQueryMeta(resp, qm)
	qm.RequestTime = rtt

	var entries []*ACLPolicyListEntry
	if err := decodeBody(resp, &entries); err != nil {
		return nil, nil, err
	}
	return entries, qm, nil
}

// RulesTranslate translates the legacy rule syntax into the current syntax.
//
// Deprecated: Support for the legacy syntax translation will be removed
// when legacy ACL support is removed.
func (a *ACL) RulesTranslate(rules io.Reader) (string, error) {
	r := a.c.newRequest("POST", "/v1/acl/rules/translate")
	r.body = rules
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	qm := &QueryMeta{}
	parseQueryMeta(resp, qm)
	qm.RequestTime = rtt

	ruleBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read translated rule body: %v", err)
	}

	return string(ruleBytes), nil
}

// RulesTranslateToken translates the rules associated with the legacy syntax
// into the current syntax and returns the results.
//
// Deprecated: Support for the legacy syntax translation will be removed
// when legacy ACL support is removed.
func (a *ACL) RulesTranslateToken(tokenID string) (string, error) {
	r := a.c.newRequest("GET", "/v1/acl/rules/translate/"+tokenID)
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	qm := &QueryMeta{}
	parseQueryMeta(resp, qm)
	qm.RequestTime = rtt

	ruleBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read translated rule body: %v", err)
	}

	return string(ruleBytes), nil
}

// RoleCreate will create a new role. It is not allowed for the role parameters
// ID field to be set as this will be generated by Consul while processing the request.
func (a *ACL) RoleCreate(role *ACLRole, q *WriteOptions) (*ACLRole, *WriteMeta, error) {
	if role.ID != "" {
		return nil, nil, fmt.Errorf("Cannot specify an ID in Role Creation")
	}

	r := a.c.newRequest("PUT", "/v1/acl/role")
	r.setWriteOptions(q)
	r.obj = role
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	wm := &WriteMeta{RequestTime: rtt}
	var out ACLRole
	if err := decodeBody(resp, &out); err != nil {
		return nil, nil, err
	}

	return &out, wm, nil
}

// RoleUpdate updates a role. The ID field of the role parameter must be set to an
// existing role ID
func (a *ACL) RoleUpdate(role *ACLRole, q *WriteOptions) (*ACLRole, *WriteMeta, error) {
	if role.ID == "" {
		return nil, nil, fmt.Errorf("Must specify an ID in Role Update")
	}

	r := a.c.newRequest("PUT", "/v1/acl/role/"+role.ID)
	r.setWriteOptions(q)
	r.obj = role
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	wm := &WriteMeta{RequestTime: rtt}
	var out ACLRole
	if err := decodeBody(resp, &out); err != nil {
		return nil, nil, err
	}

	return &out, wm, nil
}

// RoleDelete deletes a role given its ID.
func (a *ACL) RoleDelete(roleID string, q *WriteOptions) (*WriteMeta, error) {
	r := a.c.newRequest("DELETE", "/v1/acl/role/"+roleID)
	r.setWriteOptions(q)
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	wm := &WriteMeta{RequestTime: rtt}
	return wm, nil
}

// RoleRead retrieves the role details (by ID). Returns nil if not found.
func (a *ACL) RoleRead(roleID string, q *QueryOptions) (*ACLRole, *QueryMeta, error) {
	r := a.c.newRequest("GET", "/v1/acl/role/"+roleID)
	r.setQueryOptions(q)
	found, rtt, resp, err := requireNotFoundOrOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	qm := &QueryMeta{}
	parseQueryMeta(resp, qm)
	qm.RequestTime = rtt

	if !found {
		return nil, qm, nil
	}

	var out ACLRole
	if err := decodeBody(resp, &out); err != nil {
		return nil, nil, err
	}

	return &out, qm, nil
}

// RoleReadByName retrieves the role details (by name). Returns nil if not found.
func (a *ACL) RoleReadByName(roleName string, q *QueryOptions) (*ACLRole, *QueryMeta, error) {
	r := a.c.newRequest("GET", "/v1/acl/role/name/"+url.QueryEscape(roleName))
	r.setQueryOptions(q)
	found, rtt, resp, err := requireNotFoundOrOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	qm := &QueryMeta{}
	parseQueryMeta(resp, qm)
	qm.RequestTime = rtt

	if !found {
		return nil, qm, nil
	}

	var out ACLRole
	if err := decodeBody(resp, &out); err != nil {
		return nil, nil, err
	}

	return &out, qm, nil
}

// RoleList retrieves a listing of all roles. The listing does not include some
// metadata for the role as those should be retrieved by subsequent calls to
// RoleRead.
func (a *ACL) RoleList(q *QueryOptions) ([]*ACLRole, *QueryMeta, error) {
	r := a.c.newRequest("GET", "/v1/acl/roles")
	r.setQueryOptions(q)
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	qm := &QueryMeta{}
	parseQueryMeta(resp, qm)
	qm.RequestTime = rtt

	var entries []*ACLRole
	if err := decodeBody(resp, &entries); err != nil {
		return nil, nil, err
	}
	return entries, qm, nil
}

// IdentityProviderCreate will create a new identity provider.
func (a *ACL) IdentityProviderCreate(idp *ACLIdentityProvider, q *WriteOptions) (*ACLIdentityProvider, *WriteMeta, error) {
	if idp.Name == "" {
		return nil, nil, fmt.Errorf("Must specify a Name in Identity Provider Creation")
	}

	r := a.c.newRequest("PUT", "/v1/acl/idp")
	r.setWriteOptions(q)
	r.obj = idp
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	wm := &WriteMeta{RequestTime: rtt}
	var out ACLIdentityProvider
	if err := decodeBody(resp, &out); err != nil {
		return nil, nil, err
	}

	return &out, wm, nil
}

// IdentityProviderUpdate updates an identity provider.
func (a *ACL) IdentityProviderUpdate(idp *ACLIdentityProvider, q *WriteOptions) (*ACLIdentityProvider, *WriteMeta, error) {
	if idp.Name == "" {
		return nil, nil, fmt.Errorf("Must specify a Name in Identity Provider Update")
	}

	r := a.c.newRequest("PUT", "/v1/acl/idp/"+url.QueryEscape(idp.Name))
	r.setWriteOptions(q)
	r.obj = idp
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	wm := &WriteMeta{RequestTime: rtt}
	var out ACLIdentityProvider
	if err := decodeBody(resp, &out); err != nil {
		return nil, nil, err
	}

	return &out, wm, nil
}

// IdentityProviderDelete deletes an identity provider given its Name.
func (a *ACL) IdentityProviderDelete(idpName string, q *WriteOptions) (*WriteMeta, error) {
	if idpName == "" {
		return nil, fmt.Errorf("Must specify a Name in Identity Provider Delete")
	}

	r := a.c.newRequest("DELETE", "/v1/acl/idp/"+url.QueryEscape(idpName))
	r.setWriteOptions(q)
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	wm := &WriteMeta{RequestTime: rtt}
	return wm, nil
}

// IdentityProviderRead retrieves the identity provider. Returns nil if not found.
func (a *ACL) IdentityProviderRead(idpName string, q *QueryOptions) (*ACLIdentityProvider, *QueryMeta, error) {
	if idpName == "" {
		return nil, nil, fmt.Errorf("Must specify a Name in Identity Provider Read")
	}

	r := a.c.newRequest("GET", "/v1/acl/idp/"+url.QueryEscape(idpName))
	r.setQueryOptions(q)
	found, rtt, resp, err := requireNotFoundOrOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	qm := &QueryMeta{}
	parseQueryMeta(resp, qm)
	qm.RequestTime = rtt

	if !found {
		return nil, qm, nil
	}

	var out ACLIdentityProvider
	if err := decodeBody(resp, &out); err != nil {
		return nil, nil, err
	}

	return &out, qm, nil
}

// IdentityProviderList retrieves a listing of all identity providers. The
// listing does not include some metadata for the identity provider as those
// should be retrieved by subsequent calls to IdentityProviderRead.
func (a *ACL) IdentityProviderList(q *QueryOptions) ([]*ACLIdentityProviderListEntry, *QueryMeta, error) {
	r := a.c.newRequest("GET", "/v1/acl/idps")
	r.setQueryOptions(q)
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	qm := &QueryMeta{}
	parseQueryMeta(resp, qm)
	qm.RequestTime = rtt

	var entries []*ACLIdentityProviderListEntry
	if err := decodeBody(resp, &entries); err != nil {
		return nil, nil, err
	}
	return entries, qm, nil
}

// RoleBindingRuleCreate will create a new role binding rule. It is not allowed
// for the role binding rule parameter's ID field to be set as this will be
// generated by Consul while processing the request.
func (a *ACL) RoleBindingRuleCreate(rule *ACLRoleBindingRule, q *WriteOptions) (*ACLRoleBindingRule, *WriteMeta, error) {
	if rule.ID != "" {
		return nil, nil, fmt.Errorf("Cannot specify an ID in Role Binding Rule Creation")
	}

	r := a.c.newRequest("PUT", "/v1/acl/rolebindingrule")
	r.setWriteOptions(q)
	r.obj = rule
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	wm := &WriteMeta{RequestTime: rtt}
	var out ACLRoleBindingRule
	if err := decodeBody(resp, &out); err != nil {
		return nil, nil, err
	}

	return &out, wm, nil
}

// RoleBindingRuleUpdate updates a role binding rule. The ID field of the role
// binding rule parameter must be set to an existing role binding rule ID.
func (a *ACL) RoleBindingRuleUpdate(rule *ACLRoleBindingRule, q *WriteOptions) (*ACLRoleBindingRule, *WriteMeta, error) {
	if rule.ID == "" {
		return nil, nil, fmt.Errorf("Must specify an ID in Role Binding Rule Update")
	}

	r := a.c.newRequest("PUT", "/v1/acl/rolebindingrule/"+rule.ID)
	r.setWriteOptions(q)
	r.obj = rule
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	wm := &WriteMeta{RequestTime: rtt}
	var out ACLRoleBindingRule
	if err := decodeBody(resp, &out); err != nil {
		return nil, nil, err
	}

	return &out, wm, nil
}

// RoleBindingRuleDelete deletes a role binding rule given its ID.
func (a *ACL) RoleBindingRuleDelete(roleBindingRuleID string, q *WriteOptions) (*WriteMeta, error) {
	r := a.c.newRequest("DELETE", "/v1/acl/rolebindingrule/"+roleBindingRuleID)
	r.setWriteOptions(q)
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	wm := &WriteMeta{RequestTime: rtt}
	return wm, nil
}

// RoleBindingRuleRead retrieves the role binding rule details. Returns nil if not found.
func (a *ACL) RoleBindingRuleRead(roleBindingRuleID string, q *QueryOptions) (*ACLRoleBindingRule, *QueryMeta, error) {
	r := a.c.newRequest("GET", "/v1/acl/rolebindingrule/"+roleBindingRuleID)
	r.setQueryOptions(q)
	found, rtt, resp, err := requireNotFoundOrOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	qm := &QueryMeta{}
	parseQueryMeta(resp, qm)
	qm.RequestTime = rtt

	if !found {
		return nil, qm, nil
	}

	var out ACLRoleBindingRule
	if err := decodeBody(resp, &out); err != nil {
		return nil, nil, err
	}

	return &out, qm, nil
}

// RoleBindingRuleList retrieves a listing of all role binding rules.
func (a *ACL) RoleBindingRuleList(idpName string, q *QueryOptions) ([]*ACLRoleBindingRule, *QueryMeta, error) {
	r := a.c.newRequest("GET", "/v1/acl/rolebindingrules")
	if idpName != "" {
		r.params.Set("idp", idpName)
	}
	r.setQueryOptions(q)
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	qm := &QueryMeta{}
	parseQueryMeta(resp, qm)
	qm.RequestTime = rtt

	var entries []*ACLRoleBindingRule
	if err := decodeBody(resp, &entries); err != nil {
		return nil, nil, err
	}
	return entries, qm, nil
}

// Login is used to exchange identity provider credentials for a newly-minted
// Consul Token.
func (a *ACL) Login(auth *ACLLoginParams, q *WriteOptions) (*ACLToken, *WriteMeta, error) {
	r := a.c.newRequest("POST", "/v1/acl/login")
	r.setWriteOptions(q)
	r.obj = auth

	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	wm := &WriteMeta{RequestTime: rtt}
	var out ACLToken
	if err := decodeBody(resp, &out); err != nil {
		return nil, nil, err
	}
	return &out, wm, nil
}

// Logout is used to destroy a Consul Token created via Login().
func (a *ACL) Logout(q *WriteOptions) (*WriteMeta, error) {
	r := a.c.newRequest("POST", "/v1/acl/logout")
	r.setWriteOptions(q)
	rtt, resp, err := requireOK(a.c.doRequest(r))
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	wm := &WriteMeta{RequestTime: rtt}
	return wm, nil
}
