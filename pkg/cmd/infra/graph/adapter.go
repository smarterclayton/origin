package graph

import (
	"bytes"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/golang/glog"
	"github.com/gonum/graph"
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"

	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/util/sets"

	apigraph "github.com/openshift/origin/pkg/api/graph"
)

type objectNode interface {
	Object() interface{}
	Found() bool
	Kind() string
}

type imageNode interface {
	ImageSpec() string
	Kind() string
}

type objectEdge interface {
	Kinds() sets.String
}

var mergeNode = heredoc.Doc(`
	MERGE (n%[1]d:%[2]s { name: %[3]q, namespace: %[4]q } )
	ON CREATE SET n%[1]d.rv = %[5]q
	ON MATCH SET n%[1]d.rv = %[5]q, n%[1]d.deletedAt = NULL
	`)

func applyGraph(g apigraph.Graph, conn bolt.Conn) error {
	buf := &bytes.Buffer{}
	for _, n := range g.Nodes() {
		o, ok := n.(objectNode)
		if !ok || !o.Found() {
			continue
		}
		meta, err := meta.Accessor(o.Object())
		if err != nil {
			continue
		}
		fmt.Fprintf(buf, mergeNode, n.ID(), o.Kind(), meta.GetName(), meta.GetNamespace(), meta.GetResourceVersion())
	}
	for i, e := range g.Edges() {
		resourceVersion := edgeResourceVersion(e)
		if len(resourceVersion) == 0 {
			glog.V(3).Infof("Cannot find resource version for edge, must skip: %#v", e)
			continue
		}
		from := nodeToReference(e.From())
		to := nodeToReference(e.To())
		for _, kind := range g.EdgeKinds(e).List() {
			fmt.Fprintf(buf, "MERGE (%s)-[%s]->(%s)\n",
				from,
				fmt.Sprintf("r%[1]d:%[2]s { rv: %[3]q }", i, kind, resourceVersion),
				to,
			)
		}
		continue
	}
	_, err := conn.ExecNeo(buf.String(), nil)
	return err
}

func edgeResourceVersion(edge graph.Edge) string {
	if o, ok := edge.From().(objectNode); ok && o.Found() {
		if meta, err := meta.Accessor(o.Object()); err == nil {
			return meta.GetResourceVersion()
		}
	}
	if o, ok := edge.To().(objectNode); ok && o.Found() {
		if meta, err := meta.Accessor(o.Object()); err == nil {
			return meta.GetResourceVersion()
		}
	}
	return ""
}

func nodeToReference(node graph.Node) string {
	switch o := node.(type) {
	case objectNode:
		if o.Found() {
			return fmt.Sprintf("n%d", node.ID())
		}
		if meta, err := meta.Accessor(o.Object()); err == nil {
			return fmt.Sprintf("n%d:%s { name: %q, namespace: %q }", node.ID(), o.Kind(), meta.GetName(), meta.GetNamespace())
		}
	case imageNode:
		return fmt.Sprintf("n%d:%s { spec: %q }", node.ID(), o.Kind(), o.ImageSpec())
	}
	return fmt.Sprintf("n%d", node.ID())
}
