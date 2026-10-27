package leaderelection

import (
	"context"
	"errors"
	"fmt"
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
	common, err := prepareConfig(cfg.Config)
	if err != nil {
		return nil, err
	}
	cfg.Config = common
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

	electionCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	taskErr := make(chan error, 1)
	elector, err := k8sleaderelection.NewLeaderElector(k8sleaderelection.LeaderElectionConfig{
		Lock:            lock,
		LeaseDuration:   cfg.LeaseDuration,
		RenewDeadline:   cfg.RenewDeadline,
		RetryPeriod:     cfg.RetryPeriod,
		ReleaseOnCancel: false,
		Name:            cfg.LockName,
		Callbacks: k8sleaderelection.LeaderCallbacks{
			OnStartedLeading: func(leaderCtx context.Context) {
				log.Info().Str("identity", cfg.Identity).Msg("Acquired Kubernetes leadership")
				taskErr <- task(leaderCtx)
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

	select {
	case err := <-taskErr:
		if err != nil {
			return err
		}
		return errors.New("leader task stopped unexpectedly")
	case <-electionDone:
		if ctx.Err() != nil {
			return nil
		}
		return errLeadershipLost
	case <-ctx.Done():
		cancel()
		<-electionDone
		return nil
	}
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
