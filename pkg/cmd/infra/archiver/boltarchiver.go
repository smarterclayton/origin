package archiver

import (
	"github.com/boltdb/bolt"
)

type BoltArchiver struct {
}

func (a *BoltArchiver) Create(resource, namespace, name string, current *etcd.Node) error {
	return nil
}

func (a *BoltArchiver) Update(resource, namespace, name string, current *etcd.Node, previous *etcd.Node) error {
	return nil
}

func (a *BoltArchiver) Delete(resource, namespace, name string, previous *etcd.Node) error {
	return nil
}
