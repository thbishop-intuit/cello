package main

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/cello-proj/cello/internal/types"
	"github.com/cello-proj/cello/service/internal/credentials"
	"github.com/google/go-cmp/cmp"
	vault "github.com/hashicorp/vault/api"
)

const TestRole = "testRole"

var errTest = fmt.Errorf("error")

func TestVaultCreateProject(t *testing.T) {
	tests := []struct {
		name                   string
		admin                  bool
		expectedRole           string
		expectedSecret         string
		expectedSecretAccessor string
		vaultErr               error
		errResult              bool
	}{
		{
			name:                   "create project success",
			admin:                  true,
			expectedSecret:         "test-secret",
			expectedSecretAccessor: "test-secret-accessor",
			expectedRole:           "test-role",
		},
		{
			name:      "create project admin error",
			admin:     false,
			errResult: true,
		},
		{
			name:      "create project error",
			admin:     true,
			vaultErr:  errTest,
			errResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := TestRole
			if tt.admin {
				role = authorizationKeyAdmin
			}

			v := VaultProvider{
				roleID: role,
				vaultSvcFn: mockVaultSvc(vaultSvc{
					roleID: role,
					vaultLogicalSvc: &mockVaultLogical{err: tt.vaultErr, data: map[string]interface{}{
						"secret_id":          tt.expectedSecret,
						"secret_id_accessor": tt.expectedSecretAccessor,
						"role_id":            tt.expectedRole,
						"creation_time":      "2022-07-01T14:56:10.341066-07:00",
						"expiration_time":    "2023-07-01T14:56:10.341066-07:00",
					}},
					vaultSysSvc: &mockVaultSys{},
				}),
			}

			resp, err := v.CreateProject(credentials.CreateProjectArgs{
				ProjectName: "testProject",
			})

			if err != nil {
				if !tt.errResult {
					t.Errorf("\ndid not expect error, got: %v", err)
				}
			} else {
				token := resp.Token
				if tt.errResult {
					t.Errorf("\nexpected error")
				}
				if !cmp.Equal(token.RoleID, tt.expectedRole) {
					t.Errorf("\nwant: %v\n got: %v", tt.expectedRole, token.RoleID)
				}
				if !cmp.Equal(token.Secret, tt.expectedSecret) {
					t.Errorf("\nwant: %v\n got: %v", tt.expectedSecret, token.Secret)
				}
			}
		})
	}
}

func TestVaultCreateTarget(t *testing.T) {
	tests := []struct {
		name      string
		admin     bool
		vaultErr  error
		errResult bool
	}{
		{
			name:  "create target success",
			admin: true,
		},
		{
			name:      "create target admin error",
			admin:     false,
			errResult: true,
		},
		{
			name:      "create target error",
			admin:     true,
			vaultErr:  errTest,
			errResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := TestRole
			if tt.admin {
				role = authorizationKeyAdmin
			}

			v := VaultProvider{
				roleID: role,
				vaultSvcFn: mockVaultSvc(vaultSvc{
					roleID:          role,
					vaultLogicalSvc: &mockVaultLogical{err: tt.vaultErr},
				}),
			}

			_, err := v.CreateTarget(credentials.CreateTargetArgs{
				ProjectName: "testProject",
				Target:      types.Target{},
			})

			if err != nil {
				if !tt.errResult {
					t.Errorf("\ndid not expect error, got: %v", err)
				}
			} else {
				if tt.errResult {
					t.Errorf("\nexpected error")
				}
			}
		})
	}
}

func TestGetTarget(t *testing.T) {
	tests := []struct {
		name      string
		admin     bool
		vaultErr  error
		errResult bool
	}{
		{
			name:  "get target success",
			admin: true,
		},
		{
			name:      "get target admin error",
			admin:     false,
			errResult: true,
		},
		{
			name:      "get target error",
			admin:     true,
			vaultErr:  errTest,
			errResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := TestRole
			if tt.admin {
				role = authorizationKeyAdmin
			}

			v := VaultProvider{
				roleID: role,
				vaultSvcFn: mockVaultSvc(vaultSvc{
					roleID: role,
					vaultLogicalSvc: &mockVaultLogical{err: tt.vaultErr, data: map[string]interface{}{
						"role_arns":       []interface{}{"test-role-arn"},
						"policy_arns":     []interface{}{"test-policy-arn"},
						"policy_document": `{ "Version": "2012-10-17", "Statement": [ { "Effect": "Allow", "Action": "s3:ListBuckets", "Resource": "*" } ] }`,
						"credential_type": "test-cred-type",
					}},
					vaultSysSvc: &mockVaultSys{},
				}),
			}

			// TODO should this test the result?
			_, err := v.GetTarget(credentials.GetTargetArgs{
				ProjectName: "testProject",
				TargetName:  "testTarget",
			})
			if err != nil {
				if !tt.errResult {
					t.Errorf("\ndid not expect error, got: %v", err)
				}
			} else {
				if tt.errResult {
					t.Errorf("\nexpected error")
				}
			}
		})
	}
}

func TestVaultListTargets(t *testing.T) {
	tests := []struct {
		name  string
		admin bool
		// want         []string
		want            credentials.ListTargetsResponse
		expectedTargets []string
		vaultErr        error
		errResult       bool
	}{
		{
			name:  "list target success",
			admin: true,
			want: credentials.ListTargetsResponse{
				Targets: []string{"target1", "target2"},
			},
		},
		{
			name:      "list target admin error",
			admin:     false,
			errResult: true,
		},
		{
			name:      "list target error",
			admin:     true,
			vaultErr:  errTest,
			errResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := TestRole
			if tt.admin {
				role = authorizationKeyAdmin
			}

			var testTargets []interface{}
			for _, i := range tt.want.Targets {
				testTargets = append(testTargets, fmt.Sprintf("argo-cloudops-projects-test-target-%s", i))
			}

			v := VaultProvider{
				roleID: role,
				vaultSvcFn: mockVaultSvc(vaultSvc{
					roleID: role,
					vaultLogicalSvc: &mockVaultLogical{
						err: tt.vaultErr,
						data: map[string]interface{}{
							"keys": testTargets,
						},
					},
				}),
			}

			resp, err := v.ListTargets(credentials.ListTargetsArgs{
				ProjectName: "test",
			})
			if err != nil {
				if !tt.errResult {
					t.Errorf("\ndid not expect error, got: %v", err)
				}
			} else {
				if tt.errResult {
					t.Errorf("\nexpected error")
				}
				if !cmp.Equal(resp, tt.want) {
					t.Errorf("\nwant: %v\n got: %v", tt.want, resp)
				}
			}
		})
	}
}

func TestVaultProjectExists(t *testing.T) {
	tests := []struct {
		name      string
		exists    bool
		vaultErr  error
		expectErr bool
	}{
		{
			name:   "get project success",
			exists: true,
		},
		{
			name:      "get project not found",
			exists:    false,
			vaultErr:  ErrNotFound,
			expectErr: false,
		},
		{
			name:      "vault error",
			exists:    false,
			vaultErr:  errTest,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := VaultProvider{
				vaultSvcFn: mockVaultSvc(vaultSvc{
					vaultLogicalSvc: &mockVaultLogical{err: tt.vaultErr},
				}),
			}

			resp, err := v.ProjectExists(credentials.ProjectExistsArgs{
				ProjectName: "testProject",
			})
			if err != nil {
				if !tt.expectErr {
					t.Errorf("\ndid not expect error, got: %v", err)
				}
			} else {
				if tt.expectErr {
					t.Errorf("\nexpected error")
				}

				if !cmp.Equal(resp.Exists, tt.exists) {
					t.Errorf("\nwant: %v\n got: %v", tt.exists, resp.Exists)
				}
			}
		})
	}
}

func mockVaultSvc(vSvc vaultSvc) func(auth credentials.Authorization, h http.Header) (vaultSvc, error) {
	return func(auth credentials.Authorization, h http.Header) (vaultSvc, error) {
		return vSvc, nil
	}
}

type mockVaultLogical struct {
	vault.Logical
	data  map[string]interface{}
	token string
	err   error
}

func (m mockVaultLogical) Read(path string) (*vault.Secret, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &vault.Secret{Data: m.data}, nil
}

func (m mockVaultLogical) List(path string) (*vault.Secret, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &vault.Secret{Data: m.data}, nil
}

func (m mockVaultLogical) Write(path string, data map[string]interface{}) (*vault.Secret, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &vault.Secret{Data: m.data, Auth: &vault.SecretAuth{ClientToken: m.token}}, nil
}

func (m mockVaultLogical) Delete(path string) (*vault.Secret, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &vault.Secret{}, nil
}

type mockVaultSys struct {
	vault.Sys
	err error
}

func (m mockVaultSys) PutPolicy(name, rules string) error {
	return m.err
}

func (m mockVaultSys) DeletePolicy(name string) error {
	return m.err
}
