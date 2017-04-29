package main

import (
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/mitchellh/mapstructure"
	"github.com/zrepl/zrepl/rpc"
	"github.com/zrepl/zrepl/sshbytestream"
	"github.com/zrepl/zrepl/zfs"
	yaml "gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"strings"
)

type Pool struct {
	Name      string
	Transport Transport
}

type Transport interface {
	Connect() (rpc.RPCRequester, error)
}
type LocalTransport struct {
	Pool string
}
type SSHTransport struct {
	ZreplIdentity        string
	Host                 string
	User                 string
	Port                 uint16
	TransportOpenCommand []string
	Options              []string
}

type Push struct {
	To       *Pool
	Datasets []zfs.DatasetPath
}
type Pull struct {
	From    *Pool
	Mapping zfs.DatasetMapping
}
type Sink struct {
	From    string
	Mapping zfs.DatasetMapping
}

type Config struct {
	Pools []Pool
	Pushs []Push
	Pulls []Pull
	Sinks []Sink
}

func ParseConfig(path string) (config Config, err error) {

	c := make(map[string]interface{}, 0)

	var bytes []byte

	if bytes, err = ioutil.ReadFile(path); err != nil {
		return
	}

	if err = yaml.Unmarshal(bytes, &c); err != nil {
		return
	}

	return parseMain(c)
}

func parseMain(root map[string]interface{}) (c Config, err error) {
	if c.Pools, err = parsePools(root["pools"]); err != nil {
		return
	}

	poolLookup := func(name string) (*Pool, error) {
		for _, pool := range c.Pools {
			if pool.Name == name {
				return &pool, nil
			}
		}
		return nil, errors.New(fmt.Sprintf("pool '%s' not defined", name))
	}

	if c.Pushs, err = parsePushs(root["pushs"], poolLookup); err != nil {
		return
	}
	if c.Pulls, err = parsePulls(root["pulls"], poolLookup); err != nil {
		return
	}
	if c.Sinks, err = parseSinks(root["sinks"]); err != nil {
		return
	}
	return
}

func parsePools(v interface{}) (pools []Pool, err error) {

	asList := make([]struct {
		Name      string
		Transport map[string]interface{}
	}, 0)
	if err = mapstructure.Decode(v, &asList); err != nil {
		return
	}

	pools = make([]Pool, len(asList))
	for i, p := range asList {
		var transport Transport
		if transport, err = parseTransport(p.Transport); err != nil {
			return
		}
		pools[i] = Pool{
			Name:      p.Name,
			Transport: transport,
		}
	}

	return
}

func parseTransport(it map[string]interface{}) (t Transport, err error) {

	if len(it) != 1 {
		err = errors.New("ambiguous transport type")
		return
	}

	for key, val := range it {
		switch key {
		case "ssh":
			t := SSHTransport{}
			if err = mapstructure.Decode(val, &t); err != nil {
				err = errors.New(fmt.Sprintf("could not parse ssh transport: %s", err))
				return nil, err
			}
			return t, nil
		case "local":
			t := LocalTransport{}
			if err = mapstructure.Decode(val, &t); err != nil {
				err = errors.New(fmt.Sprintf("could not parse local transport: %s", err))
				return nil, err
			}
			return t, nil
		default:
			return nil, errors.New(fmt.Sprintf("unknown transport type '%s'\n", key))
		}
	}

	return // unreachable

}

type poolLookup func(name string) (*Pool, error)

func parsePushs(v interface{}, pl poolLookup) (p []Push, err error) {

	asList := make([]struct {
		To       string
		Datasets []string
	}, 0)

	if err = mapstructure.Decode(v, &asList); err != nil {
		return
	}

	p = make([]Push, len(asList))

	for i, e := range asList {

		var toPool *Pool
		if toPool, err = pl(e.To); err != nil {
			return
		}
		push := Push{
			To:       toPool,
			Datasets: make([]zfs.DatasetPath, len(e.Datasets)),
		}

		for i, ds := range e.Datasets {
			if push.Datasets[i], err = zfs.NewDatasetPath(ds); err != nil {
				return
			}
		}

		p[i] = push
	}

	return
}

func parsePulls(v interface{}, pl poolLookup) (p []Pull, err error) {

	asList := make([]struct {
		From    string
		Mapping map[string]string
	}, 0)

	if err = mapstructure.Decode(v, &asList); err != nil {
		return
	}

	p = make([]Pull, len(asList))

	for i, e := range asList {

		var fromPool *Pool
		if fromPool, err = pl(e.From); err != nil {
			return
		}
		pull := Pull{
			From: fromPool,
		}
		if pull.Mapping, err = parseComboMapping(e.Mapping); err != nil {
			return
		}
		p[i] = pull
	}

	return
}

func parseSinks(v interface{}) (sinks []Sink, err error) {

	var asList []interface{}
	var ok bool
	if asList, ok = v.([]interface{}); !ok {
		return nil, errors.New("expected list")
	}

	sinks = make([]Sink, len(asList))

	for i, s := range asList {
		var sink Sink
		if sink, err = parseSink(s); err != nil {
			return
		}
		sinks[i] = sink
	}

	return
}

func parseSink(v interface{}) (s Sink, err error) {
	t := struct {
		From    string
		Mapping map[string]string
	}{}
	if err = mapstructure.Decode(v, &t); err != nil {
		return
	}

	s.From = t.From
	s.Mapping, err = parseComboMapping(t.Mapping)
	return
}

func parseComboMapping(m map[string]string) (c zfs.ComboMapping, err error) {

	c.Mappings = make([]zfs.DatasetMapping, len(m))

	for lhs, rhs := range m {

		if lhs[0] == '|' {

			if len(m) != 1 {
				err = errors.New("non-recursive mapping must be the only mapping for a sink")
			}

			m := zfs.DirectMapping{
				Source: nil,
			}

			if m.Target, err = zfs.NewDatasetPath(rhs); err != nil {
				return
			}

			c.Mappings = append(c.Mappings, m)

		} else if lhs[0] == '*' {

			m := zfs.ExecMapping{}
			fields := strings.Fields(strings.TrimPrefix(rhs, "!"))
			if len(fields) < 1 {
				err = errors.New("ExecMapping without acceptor path")
				return
			}
			m.Name = fields[0]
			m.Args = fields[1:]

			c.Mappings = append(c.Mappings, m)

		} else if strings.HasSuffix(lhs, "*") {

			m := zfs.GlobMapping{}

			m.PrefixPath, err = zfs.NewDatasetPath(strings.TrimSuffix(lhs, "*"))
			if err != nil {
				return
			}

			if m.TargetRoot, err = zfs.NewDatasetPath(rhs); err != nil {
				return
			}

			c.Mappings = append(c.Mappings, m)

		}

	}

	return

}

func (t SSHTransport) Connect() (r rpc.RPCRequester, err error) {
	var stream io.ReadWriteCloser
	var rpcTransport sshbytestream.SSHTransport
	copier.Copy(rpcTransport, t)
	if stream, err = sshbytestream.Outgoing(rpcTransport); err != nil {
		return
	}
	return rpc.ConnectByteStreamRPC(stream)
}

func (t LocalTransport) Connect() (r rpc.RPCRequester, err error) {
	// TODO ugly hidden global variable reference
	return rpc.ConnectLocalRPC(handler), nil
}
