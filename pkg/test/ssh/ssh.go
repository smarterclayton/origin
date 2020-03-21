package ssh

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/spf13/pflag"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog"
)

// Options supports an SSH server that forwards requests from users that authorize
// with public keys in AuthorizedKeys to backend addresses returned by AddressFunc.
type Options struct {
	Address         string
	ServerKey       string
	AllowedBackends []string
	AuthorizedKeys  string

	AuthorizedKeysFunc func() ([]ssh.PublicKey, error)
	AddressFunc        func(string) (string, int, bool, error)
}

func (o *Options) Bind(flags *pflag.FlagSet) {
	flags.StringVar(&o.Address, "address", o.Address, "The network address to listen for requests on.")
	flags.StringVar(&o.ServerKey, "server-key-file", o.ServerKey, "Path to the private key for this server. Will be auto generated if empty.")
	flags.StringVar(&o.AuthorizedKeys, "authorized-keys-file", o.AuthorizedKeys, "Path to a file containing public keys in authorized_keys format that may forward SSH requests.")
	flags.StringSliceVar(&o.AllowedBackends, "allowed-backends", o.AllowedBackends, "A set of allowed backends to forward to.")
}

func (o *Options) Run() error {
	config := &ssh.ServerConfig{}

	var authorizedKeys []ssh.PublicKey
	if len(o.AuthorizedKeys) > 0 {
		data, err := ioutil.ReadFile(o.AuthorizedKeys)
		if err != nil {
			return fmt.Errorf("unable to read authorized keys: %v", err)
		}
		for len(data) > 0 {
			key, _, _, rest, err := ssh.ParseAuthorizedKey(data)
			if err != nil {
				return fmt.Errorf("invalid authorized keys: %v", err)
			}
			authorizedKeys = append(authorizedKeys, key)
			data = rest
		}
	}
	keysFn := func() ([]ssh.PublicKey, error) {
		if o.AuthorizedKeysFunc == nil {
			return authorizedKeys, nil
		}
		keys, err := o.AuthorizedKeysFunc()
		return append(keys, authorizedKeys...), err
	}
	if keys, err := keysFn(); err == nil {
		for _, key := range keys {
			data := ssh.MarshalAuthorizedKey(key)
			klog.V(2).Infof("Authorized %s", string(data[:len(data)-1]))
		}
	}
	config.PublicKeyCallback = func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
		authorizedKeys, err := keysFn()
		if err != nil {
			klog.Warningf("Unable to load authorized keys: %v", err)
		} else if len(authorizedKeys) == 0 {
			klog.Warningf("No authorized keys are available, all requests will be denied")
		}
		keyBytes := key.Marshal()
		for _, candidate := range authorizedKeys {
			if candidate.Type() != key.Type() || !bytes.Equal(candidate.Marshal(), keyBytes) {
				continue
			}
			return &ssh.Permissions{}, nil
		}
		return nil, fmt.Errorf("user is not authorized")
	}

	backendsFn := o.AddressFunc
	if backendsFn == nil || len(o.AllowedBackends) > 0 {
		backendsFn = func(address string) (string, int, bool, error) {
			for _, allowed := range o.AllowedBackends {
				if allowed == address {
					return address, 22, true, nil
				}
			}
			return "", 0, false, nil
		}
	}

	if len(o.ServerKey) > 0 {
		data, err := ioutil.ReadFile(o.ServerKey)
		if err != nil {
			return fmt.Errorf("failed to read server key file: %v", err)
		}
		private, err := ssh.ParsePrivateKey(data)
		if err != nil {
			return fmt.Errorf("failed to parse server key file: %v", err)
		}
		klog.Infof("Host key: %s", ssh.FingerprintSHA256(private.PublicKey()))
		config.AddHostKey(private)
	} else {
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return fmt.Errorf("failed to generate server key file: %v", err)
		}
		private, err := ssh.NewSignerFromKey(key)
		if err != nil {
			return fmt.Errorf("failed to initialize server key file: %v", err)
		}
		klog.Infof("Host key: %s", ssh.FingerprintSHA256(private.PublicKey()))
		config.AddHostKey(private)
	}

	server := &server{
		config:  config,
		dial:    (&net.Dialer{Timeout: 30 * time.Second}).Dial,
		address: backendsFn,
	}
	listenConfig := net.ListenConfig{}

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	listener, err := listenConfig.Listen(ctx, "tcp", o.Address)
	if err != nil {
		return err
	}
	klog.Infof("Listening for SSH proxy traffic on %s", o.Address)
	return server.Serve(listener)
}

type server struct {
	config *ssh.ServerConfig
	dial   func(address, protocol string) (net.Conn, error)

	lock    sync.Mutex
	address func(address string) (string, int, bool, error)
}

func (s *server) Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go func() {
			if err := s.handleConn(conn); err != nil {
				klog.Errorf("[%s->] terminated: %v", conn.RemoteAddr().String(), err)
			}
		}()
	}
}

func (s *server) allowConnection(conn *ssh.ServerConn) error {
	if conn.Permissions == nil {
		return fmt.Errorf("authorization rejected")
	}
	return nil
}

func (s *server) handleConn(c net.Conn) error {
	conn, chans, reqs, err := ssh.NewServerConn(c, s.config)
	if err != nil {
		c.Close()
		return err
	}
	defer conn.Close()

	if err := s.allowConnection(conn); err != nil {
		return err
	}

	klog.V(4).Infof("[%s->] accepted", conn.RemoteAddr().String())

	go func() {
		for req := range reqs {
			switch req.Type {
			case "keepalive@openssh.com":
				if req.WantReply {
					req.Reply(true, []byte{})
				}
			default:
				req.Reply(false, []byte{})
			}
		}
	}()

	newChannel := <-chans
	if newChannel == nil {
		klog.V(4).Infof("No new channel, exiting: %s", conn.RemoteAddr().String())
		return nil
	}

	switch newChannel.ChannelType() {
	case "direct-tcpip":
		return s.ChannelForward(conn.RemoteAddr().String(), newChannel)
	default:
		newChannel.Reject(ssh.UnknownChannelType, "only channel forwarding is supported")
		return fmt.Errorf("rejected due to invalid channel type: %s", newChannel.ChannelType())
	}
}

// https://tools.ietf.org/html/rfc4254
type channelOpenDirectMsg struct {
	RAddr string
	RPort uint32
	LAddr string
	LPort uint32
}

func (s *server) ChannelForward(remoteAddr string, newChannel ssh.NewChannel) error {
	var msg channelOpenDirectMsg
	ssh.Unmarshal(newChannel.ExtraData(), &msg)

	address := fmt.Sprintf("%s:%d", msg.RAddr, msg.RPort)
	host, port, ok, err := s.address(msg.RAddr)
	if err != nil {
		newChannel.Reject(ssh.Prohibited, "remote host access denied for user")
		return fmt.Errorf("remote host access denied: %s: %v", address, err)
	}
	if !ok {
		newChannel.Reject(ssh.Prohibited, "remote host access denied for user")
		return fmt.Errorf("remote host access denied: %s", address)
	}
	target := net.JoinHostPort(host, strconv.Itoa(port))

	klog.V(4).Infof("[%s->%s] dial", remoteAddr, target)

	conn, err := s.dial("tcp", target)
	if err != nil {
		newChannel.Reject(ssh.ConnectionFailed, fmt.Sprintf("error: %v", err))
		return err
	}

	channel, reqs, err := newChannel.Accept()
	if err != nil {
		return err
	}

	go ssh.DiscardRequests(reqs)

	errCh := make(chan error, 2)
	go func() {
		defer channel.Close()
		n, err := io.Copy(channel, conn)
		klog.V(4).Infof("[%s->%s] recv complete: %d %v", remoteAddr, target, n, err)
		errCh <- err
	}()
	go func() {
		defer conn.Close()
		n, err := io.Copy(conn, channel)
		klog.V(4).Infof("[%s->%s] send complete: %d %v", remoteAddr, target, n, err)
		errCh <- err
	}()
	if err := <-errCh; err != nil {
		// ignore the remaining error
		<-errCh
		return err
	}
	return <-errCh
}

type publicKey struct {
	publicKey     []byte
	publicKeyType string
}

func (p *publicKey) Marshal() []byte {
	b := make([]byte, len(p.publicKey))
	copy(b, p.publicKey)
	return b
}

func (p *publicKey) Type() string {
	return p.publicKeyType
}

func (p *publicKey) Verify([]byte, *ssh.Signature) error {
	return errors.New("verify not implemented")
}
