// Package archiver creates a record of the changes to the OpenShift and
// Kubernetes schema based on the underlying etcd event log. It also maintains
// secondary indices on other access patterns:
//
// * Event log - tracks creates, updates, and deletes
//   (etcdIndex, resourceType, namespace, name) -> [lastEtcdIndex, contents] (create/update) or [lastEtcdIndex] (delete)
//   Supports efficient in-order traversal of the change log
//
// * Resources by version
//   (resourceType, namespace, name, etcdIndex) -> []
//   Supports efficient retrieval of all historical versions of a resource
//
// * Resources by uid
//   (uid) -> [etcdIndex] or [etcdIndex, deletedEtcdIndex] (if deleted)
//   Supports efficient retrieval of resources by uid
//
// * Deleted resources in a namespace
//   (resourceType, namespace, deletedEtcdIndex, uid) -> [lastEtcdIndex]
//   Supports efficient retrieval of the deleted resources in a namespace in order
//
package archiver

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/boltdb/bolt"
	"github.com/coreos/go-etcd/etcd"

	"github.com/openshift/origin/pkg/archiver/tuple"
)

var bucketEvents = tuple.Key("events")

func OpenBoltArchiver(path string, mode os.FileMode) (*BoltArchiver, error) {
	db, err := bolt.Open(path, mode, nil)
	if err != nil {
		return nil, err
	}
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketEvents)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return NewBoltArchiver(db), nil
}

type BoltArchiver struct {
	db *bolt.DB
}

func NewBoltArchiver(db *bolt.DB) *BoltArchiver {
	return &BoltArchiver{db}
}

func (a *BoltArchiver) Close() error {
	return a.db.Close()
}

func (a *BoltArchiver) Dump(w io.Writer) error {
	return a.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketEvents)
		bucket.ForEach(func(k, v []byte) error {
			ks := hex.EncodeToString(k)
			vs := hex.EncodeToString(v)
			fmt.Fprintf(w, "%d[%s]=%s\n", len(k), ks, vs)
			return nil
		})
		return nil
	})
}

func (a *BoltArchiver) Create(resource, namespace, name string, index uint64, current *etcd.Node) error {
	uid, ok := extractUID(current.Value)
	return a.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketEvents)

		if err := bucket.Put(keyByIndex(resource, namespace, name, index).Pack(), tuple.Tuple{index, current.Value}.Pack()); err != nil {
			return err
		}

		if err := bucket.Put(keyByType(resource, namespace, name, index).Pack(), nil); err != nil {
			return err
		}

		if ok {
			if err := bucket.Put([]byte(uid), tuple.Tuple{index}.Pack()); err != nil {
				return err
			}
		}

		return nil
	})
}

func (a *BoltArchiver) Update(resource, namespace, name string, index uint64, current *etcd.Node, previous *etcd.Node) error {
	uid, ok := extractUID(current.Value)
	return a.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketEvents)

		if err := bucket.Put(keyByIndex(resource, namespace, name, index).Pack(), tuple.Tuple{previous.ModifiedIndex, current.Value}.Pack()); err != nil {
			return err
		}

		prefix := keyByType(resource, namespace, name, index)[:2]
		if err := deleteRange(bucket, prefix.Pack()); err != nil {
			return err
		}
		if err := bucket.Put(keyByType(resource, namespace, name, index).Pack(), nil); err != nil {
			return err
		}

		if ok {
			if err := bucket.Put([]byte(uid), tuple.Tuple{index}.Pack()); err != nil {
				return err
			}
		}

		return nil
	})
}

func (a *BoltArchiver) Delete(resource, namespace, name string, index uint64, previous *etcd.Node) error {
	uid, ok := extractUID(previous.Value)
	return a.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketEvents)

		if err := bucket.Put(keyByIndex(resource, namespace, name, index).Pack(), tuple.Tuple{previous.ModifiedIndex}.Pack()); err != nil {
			return err
		}

		prefix := keyByType(resource, namespace, name, index)[:2]
		if err := deleteRange(bucket, prefix.Pack()); err != nil {
			return err
		}

		if ok {
			if err := bucket.Delete([]byte(uid)); err != nil {
				return err
			}
		}

		return nil
	})
}

func deleteRange(bucket *bolt.Bucket, prefix tuple.Key) error {
	cursor := bucket.Cursor()
	for k, _ := cursor.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = cursor.Next() {
		if err := bucket.Delete(k); err != nil {
			return err
		}
	}
	return nil
}

func keyByIndex(resource, namespace, name string, index uint64) tuple.Tuple {
	return tuple.Tuple{index, resource, namespace, name}
}

func keyByType(resource, namespace, name string, index uint64) tuple.Tuple {
	return tuple.Tuple{resource, namespace, name, index}
}

type uidMetadata struct {
	UID string `json:"uid"`
}

type uidObject struct {
	Metadata uidMetadata `json:"metadata"`

	UID string `json:"uid"`
}

func extractUID(value string) (string, bool) {
	obj := uidObject{}
	if err := json.Unmarshal([]byte(value), &obj); err != nil {
		return "", false
	}
	if len(obj.Metadata.UID) > 0 {
		return obj.Metadata.UID, true
	}
	if len(obj.UID) > 0 {
		return obj.UID, true
	}
	return "", false
}
