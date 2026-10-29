package leaderelection

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	k8sleaderelection "k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

// KubernetesConfig configures Kubernetes Lease-based leader election.
type KubernetesConfig struct {
	Config
	Namespace     string
	LeaseDuration time.Duration
	RenewDeadline time.Duration
}

// Validate checks the Kubernetes leader-election configuration.
func (cfg KubernetesConfig) Validate() error {
	if err := cfg.Config.Validate(); err != nil {
		return err
	}
	if cfg.Namespace == "" {
		return errors.New("leader election namespace must not be empty")
	}
	if cfg.LeaseDuration <= cfg.RenewDeadline {
		return errors.New("leader election lease duration must be greater than renew deadline")
	}
	if cfg.RenewDeadline <= cfg.RetryPeriod {
		return errors.New("leader election renew deadline must be greater than retry period")
	}
	return nil
}

type kubernetesLeader struct {
	cfg    KubernetesConfig
	client kubernetes.Interface
}

var _ Leader = (*kubernetesLeader)(nil)

// NewKubernetes constructs a Kubernetes leader using kubeconfig or in-cluster credentials.
func NewKubernetes(kubeconfig string, cfg KubernetesConfig) (Leader, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	sharedConfig, err := prepareConfig(cfg.Config)
	if err != nil {
		return nil, err
	}
	cfg.Config = sharedConfig
	restConfig, err := kubernetesConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("create Kubernetes client: %w", err)
	}
	return &kubernetesLeader{cfg: cfg, client: client}, nil
}

func (l *kubernetesLeader) Lead(ctx context.Context, task Task) error {
	cfg := l.cfg
	client := l.client
	lock, err := resourcelock.New(
		resourcelock.LeasesResourceLock,
		cfg.Namespace,
		cfg.LockName,
		client.CoreV1(),
		client.CoordinationV1(),
		resourcelock.ResourceLockConfig{Identity: cfg.Identity},
	)
	if err != nil {
		return fmt.Errorf("create Kubernetes Lease lock: %w", err)
	}

	// Keep renewing during graceful task shutdown; release only after the task exits.
	electionCtx, stopElection := context.WithCancel(context.WithoutCancel(ctx))
	defer stopElection()
	taskCtx, cancelTask := context.WithCancel(ctx)
	defer cancelTask()
	var taskStartupGate sync.Once
	var taskWG sync.WaitGroup
	taskResult := make(chan error, 1)
	stopTaskAndWait := func() {
		// Close the startup gate, or wait for task registration if startup won.
		taskStartupGate.Do(func() {})
		cancelTask()
		taskWG.Wait()
	}
	lock = &taskShutdownLock{Interface: lock, stopTaskAndWait: stopTaskAndWait}
	elector, err := k8sleaderelection.NewLeaderElector(k8sleaderelection.LeaderElectionConfig{
		Lock:            lock,
		LeaseDuration:   cfg.LeaseDuration,
		RenewDeadline:   cfg.RenewDeadline,
		RetryPeriod:     cfg.RetryPeriod,
		ReleaseOnCancel: true,
		Name:            cfg.LockName,
		Callbacks: k8sleaderelection.LeaderCallbacks{
			OnStartedLeading: func(leaderCtx context.Context) {
				taskStartupGate.Do(func() {
					taskWG.Go(func() {
						unlinkLeadershipCancellation := context.AfterFunc(leaderCtx, cancelTask)
						defer unlinkLeadershipCancellation()
						log.Info().Str("identity", cfg.Identity).Msg("Acquired Kubernetes leadership")
						taskResult <- task(taskCtx)
					})
				})
			},
			OnStoppedLeading: func() {
				log.Warn().Str("identity", cfg.Identity).Msg("Lost Kubernetes leadership")
			},
			OnNewLeader: func(identity string) {
				if identity != cfg.Identity {
					log.Info().Str("identity", identity).Msg("Observed new Kubernetes leader")
				}
			},
		},
	})
	if err != nil {
		return fmt.Errorf("create Kubernetes leader elector: %w", err)
	}

	electionDone := make(chan struct{})
	go func() {
		elector.Run(electionCtx)
		close(electionDone)
	}()

	var result error
	select {
	case err := <-taskResult:
		if ctx.Err() == nil {
			result = err
			if result == nil {
				result = errors.New("leader task stopped unexpectedly")
			}
		}
	case <-electionDone:
		if ctx.Err() == nil {
			result = errLeadershipLost
		}
	case <-ctx.Done():
	}

	stopTaskAndWait()
	stopElection()
	<-electionDone
	return result
}

// taskShutdownLock lets client-go release the Lease only after work has stopped,
// including when renewal fails before the election context is canceled.
type taskShutdownLock struct {
	resourcelock.Interface
	stopTaskAndWait func()
}

func (l *taskShutdownLock) Update(ctx context.Context, record resourcelock.LeaderElectionRecord) error {
	if record.HolderIdentity == "" {
		l.stopTaskAndWait()
	}
	return l.Interface.Update(ctx, record)
}

func kubernetesConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("load kubeconfig: %w", err)
		}
		return cfg, nil
	}
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("load in-cluster Kubernetes configuration: %w", err)
	}
	return cfg, nil
}
