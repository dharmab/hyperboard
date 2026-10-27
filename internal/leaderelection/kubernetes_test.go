package leaderelection

import (
	"context"
	"errors"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKubernetesShutdownWaitsForTaskBeforeReleasingLease(t *testing.T) {
	t.Parallel()
	for _, taskFails := range []bool{false, true} {
		name := "cancellation"
		if taskFails {
			name = "task failure"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			client := fake.NewClientset()
			leader := &kubernetesLeader{
				client: client,
				cfg: KubernetesConfig{
					Config:    Config{Identity: "test", LockName: "test", RetryPeriod: 10 * time.Millisecond},
					Namespace: "default", LeaseDuration: 3 * time.Second, RenewDeadline: time.Second,
				},
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			started := make(chan struct{})
			stopping := make(chan struct{})
			finish := make(chan struct{})
			defer func() {
				select {
				case <-finish:
				default:
					close(finish)
				}
			}()
			taskFailure := errors.New("task failed")
			done := make(chan error, 1)
			go func() {
				done <- leader.Lead(ctx, func(taskCtx context.Context) error {
					close(started)
					if taskFails {
						<-finish
						return taskFailure
					}
					<-taskCtx.Done()
					close(stopping)
					<-finish
					return taskCtx.Err()
				})
			}()
			select {
			case <-started:
			case <-time.After(5 * time.Second):
				t.Fatal("task did not start")
			}
			if !taskFails {
				cancel()
				select {
				case <-stopping:
				case <-time.After(5 * time.Second):
					t.Fatal("task was not canceled")
				}
				select {
				case err := <-done:
					t.Fatalf("Lead returned before task stopped: %v", err)
				default:
				}
			}
			lease, err := client.CoordinationV1().Leases("default").Get(context.Background(), "test", metav1.GetOptions{})
			require.NoError(t, err)
			require.NotNil(t, lease.Spec.HolderIdentity)
			assert.Equal(t, "test", *lease.Spec.HolderIdentity)
			close(finish)
			select {
			case err := <-done:
				if taskFails {
					require.ErrorIs(t, err, taskFailure)
				} else {
					require.NoError(t, err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("Lead did not return")
			}
			lease, err = client.CoordinationV1().Leases("default").Get(context.Background(), "test", metav1.GetOptions{})
			require.NoError(t, err)
			require.NotNil(t, lease.Spec.HolderIdentity)
			assert.Empty(t, *lease.Spec.HolderIdentity)
		})
	}
}
