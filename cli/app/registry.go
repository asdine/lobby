package app

import (
	"context"
	"path"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/bolt"
	"github.com/asdine/lobby/etcd"
	"github.com/asdine/lobby/log"
	"github.com/coreos/etcd/clientv3"
	"github.com/pkg/errors"
)

type registryStep int

func (registryStep) setup(ctx context.Context, app *App) error {
	var reg lobby.Registry
	var err error
	switch app.Config.Registry {
	case "":
		fallthrough
	case "bolt":
		app.Logger.Debug("Using bolt registry")
		reg, err = boltRegistry(ctx, app)
	case "etcd":
		app.Logger.Debug("Using etcd registry")
		reg, err = etcdRegistry(ctx, app)
	default:
		err = errors.New("unknown registry")
	}
	if err != nil {
		return err
	}

	app.registry = reg
	return nil
}

func boltRegistry(ctx context.Context, app *App) (lobby.Registry, error) {
	dataPath := path.Join(app.Config.Paths.DataDir, "db")
	err := createDir(dataPath)
	if err != nil {
		return nil, err
	}

	boltPath := path.Join(dataPath, "bolt")
	err = createDir(boltPath)
	if err != nil {
		return nil, err
	}

	registryPath := path.Join(boltPath, "registry.db")

	return bolt.NewRegistry(registryPath, log.New(log.Prefix("bolt registry:"), log.Debug(app.Config.Debug)))
}

func etcdRegistry(ctx context.Context, app *App) (lobby.Registry, error) {
	client, err := clientv3.New(app.Config.Etcd)
	if err != nil {
		return nil, err
	}

	return etcd.NewRegistry(
		client,
		log.New(log.Prefix("etcd registry:"), log.Debug(app.Config.Debug)),
		"lobby",
	)
}

func (registryStep) teardown(ctx context.Context, app *App) error {
	if app.registry != nil {
		app.Logger.Debug("Closing registry")
		err := app.registry.Close()
		app.registry = nil
		return err
	}

	return nil
}
