package generic

import (
	"context"
	"sync"
	"time"

	"github.com/rancher/lasso/pkg/cache"
	"github.com/rancher/lasso/pkg/client"
	"github.com/rancher/lasso/pkg/controller"
	"github.com/rancher/lasso/pkg/log"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/wrangler/v2/pkg/schemes"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
)

func init() {
	log.Infof = logrus.Infof
	log.Errorf = logrus.Errorf
}

type Factory struct {
	lock              sync.Mutex
	cacheFactory      cache.SharedCacheFactory
	controllerFactory controller.SharedControllerFactory
	threadiness       map[schema.GroupVersionKind]int
	config            *rest.Config
	Opts              FactoryOptions
}

type FactoryOptions struct {
	Namespace               string
	Resync                  time.Duration
	SharedCacheFactory      cache.SharedCacheFactory
	SharedControllerFactory controller.SharedControllerFactory
	HealthCallback          func(bool)
	TS                      *session.Session
}

func NewFactoryFromConfigWithOptions(config *rest.Config, opts *FactoryOptions) (*Factory, error) {
	if opts == nil {
		opts = &FactoryOptions{}
	}

	f := &Factory{
		config:            config,
		threadiness:       map[schema.GroupVersionKind]int{},
		cacheFactory:      opts.SharedCacheFactory,
		controllerFactory: opts.SharedControllerFactory,
		Opts:              *opts,
	}

	if f.cacheFactory == nil && f.controllerFactory != nil {
		f.cacheFactory = f.controllerFactory.SharedCacheFactory()
	}

	return f, nil
}

func (c *Factory) SetThreadiness(gvk schema.GroupVersionKind, threadiness int) {
	c.threadiness[gvk] = threadiness
}

func (c *Factory) ControllerFactory() controller.SharedControllerFactory {
	err := c.setControllerFactoryWithLock()
	utilruntime.Must(err)
	return c.controllerFactory
}

func (c *Factory) setControllerFactoryWithLock() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.controllerFactory != nil {
		return nil
	}

	cacheFactory := c.cacheFactory
	if cacheFactory == nil {
		client, err := client.NewSharedClientFactory(c.config, &client.SharedClientFactoryOptions{
			Scheme: schemes.All,
		})
		if err != nil {
			return err
		}

		cacheFactory = cache.NewSharedCachedFactory(client, &cache.SharedCacheFactoryOptions{
			DefaultNamespace: c.Opts.Namespace,
			DefaultResync:    c.Opts.Resync,
			HealthCallback:   c.Opts.HealthCallback,
		})
	}

	c.cacheFactory = cacheFactory
	c.controllerFactory = controller.NewSharedControllerFactory(cacheFactory, &controller.SharedControllerFactoryOptions{
		KindWorkers: c.threadiness,
	})

	return nil
}

func (c *Factory) Sync(ctx context.Context) error {
	if c.cacheFactory != nil {
		_ = c.cacheFactory.Start(ctx)
		c.cacheFactory.WaitForCacheSync(ctx)
	}
	return nil
}

func (c *Factory) Start(ctx context.Context, defaultThreadiness int) error {
	if err := c.Sync(ctx); err != nil {
		return err
	}

	if c.controllerFactory != nil {
		return c.controllerFactory.Start(ctx, defaultThreadiness)
	}

	return nil
}
