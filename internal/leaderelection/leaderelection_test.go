package leaderelection

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPostgres(t *testing.T) {
	t.Parallel()
	cfg := Config{LockName: "test", RetryPeriod: time.Second}
	leader, err := NewPostgres(nil, cfg)
	require.NoError(t, err)
	identity, err := os.Hostname()
	require.NoError(t, err)
	assert.Equal(t, identity, leader.(*postgresLeader).cfg.Identity)

	cfg.Identity = "explicit-identity"
	leader, err = NewPostgres(nil, cfg)
	require.NoError(t, err)
	assert.Equal(t, cfg.Identity, leader.(*postgresLeader).cfg.Identity)
}

func TestNewKubernetes(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "kubeconfig")
	require.NoError(t, os.WriteFile(path, []byte(`apiVersion: v1
kind: Config
clusters:
- name: test
  cluster:
    server: https://localhost:6443
contexts:
- name: test
  context:
    cluster: test
    user: test
current-context: test
users:
- name: test
  user:
    token: test
`), 0o600))
	cfg := KubernetesConfig{
		LockName: "test", Identity: "test", RetryPeriod: 2 * time.Second,
		Namespace:     "default",
		LeaseDuration: 15 * time.Second, RenewDeadline: 10 * time.Second,
	}
	leader, err := NewKubernetes(path, cfg)
	require.NoError(t, err)
	assert.Equal(t, cfg, leader.(*kubernetesLeader).cfg)
	assert.NotNil(t, leader.(*kubernetesLeader).client)

	leader, err = NewKubernetes(filepath.Join(t.TempDir(), "missing"), cfg)
	require.ErrorContains(t, err, "load kubeconfig")
	assert.Nil(t, leader)

	cfg.Namespace = ""
	_, err = NewKubernetes(path, cfg)
	require.EqualError(t, err, "leader election namespace must not be empty")
}

func TestKubernetesConfigValidate(t *testing.T) {
	t.Parallel()
	valid := KubernetesConfig{
		LockName: "test", RetryPeriod: 2 * time.Second,
		Namespace: "default", LeaseDuration: 15 * time.Second, RenewDeadline: 10 * time.Second,
	}
	for _, tc := range []struct {
		name   string
		modify func(*KubernetesConfig)
		want   string
	}{
		{"lock name", func(cfg *KubernetesConfig) { cfg.LockName = "" }, "leader election lock name must not be empty"},
		{"retry period", func(cfg *KubernetesConfig) { cfg.RetryPeriod = 0 }, "leader election retry period must be positive"},
		{"namespace", func(cfg *KubernetesConfig) { cfg.Namespace = "" }, "leader election namespace must not be empty"},
		{"lease duration", func(cfg *KubernetesConfig) { cfg.LeaseDuration = cfg.RenewDeadline }, "leader election lease duration must be greater than renew deadline"},
		{"renew deadline", func(cfg *KubernetesConfig) { cfg.RenewDeadline = cfg.RetryPeriod }, "leader election renew deadline must be greater than retry period"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := valid
			tc.modify(&cfg)
			require.EqualError(t, cfg.Validate(), tc.want)
		})
	}
	require.NoError(t, valid.Validate())
}

func TestNewPostgresRejectsInvalidConfig(t *testing.T) {
	t.Parallel()
	leader, err := NewPostgres(nil, Config{})
	require.Error(t, err)
	assert.Nil(t, leader)
}
