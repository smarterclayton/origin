package etcdcmd

import (
	"fmt"
	"time"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/tools"
	"github.com/coreos/go-etcd/etcd"
	"github.com/spf13/pflag"

	"github.com/openshift/origin/pkg/cmd/flagtypes"
	"github.com/openshift/origin/pkg/cmd/util"
)

const ConfigSyntax = " --etcd=<addr>"

type Config struct {
	EtcdAddr flagtypes.Addr
}

func NewConfig() *Config {
	return &Config{
		EtcdAddr: flagtypes.Addr{Value: "localhost:4001", DefaultScheme: "https", DefaultPort: 4001, AllowPrefix: false}.Default(),
	}
}

func (cfg *Config) Bind(flag *pflag.FlagSet) {
	flag.Var(&cfg.EtcdAddr, "etcd", "The address etcd can be reached on (host, host:port, or URL).")
}

func (cfg *Config) bindEnv() {
	if value, ok := util.GetEnv("ETCD_ADDR"); ok {
		cfg.EtcdAddr.Set(value)
	}
}

func (cfg *Config) Client(check bool) (*etcd.Client, error) {
	cfg.bindEnv()

	etcdServers := []string{cfg.EtcdAddr.URL.String()}
	etcdClient := etcd.NewClient(etcdServers)

	if check {
		for i := 0; ; i += 1 {
			_, err := etcdClient.Get("/", false, false)
			if err == nil || tools.IsEtcdNotFound(err) {
				break
			}
			if i > 100 {
				return nil, fmt.Errorf("Could not reach etcd at %q: %v", cfg.EtcdAddr.URL, err)
			}
			time.Sleep(50 * time.Millisecond)
		}
	}

	return etcdClient, nil
}
