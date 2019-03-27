package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/untillpro/igoonce/iconfig"
	"io"
	"net/http"

	"github.com/untillpro/godif"
)

// Declare s.e.
func Declare() {
	godif.Provide(&iconfig.GetCurrentAppConfig, getCurrentAppConfig)
	godif.Provide(&iconfig.PutCurrentAppConfig, putCurrentAppConfig)
}

func main() {
	a := ConsulConfig{"127.0.0.1", "foo", 8500}
	c := a.putConfig(map[string]interface{}{"foo": ConsulConfig{"a", "b", 12}})
	b, c := a.getConfig()
	if c != nil {
		panic(c)
	}
	fmt.Println(b)
	if c != nil {
		panic(c)
	}
	//x := b.([]map[string]interface{})[0]["Value"]
	//xz,_ := base64.StdEncoding.DecodeString(x.(string))
	//fmt.Println(string(xz))
}

type consulKey int

const consul consulKey = 0

type ConsulConfig struct {
	host, prefix string
	port         int
}

// KVPair is used to represent a single K/V entry
type ConsulEntry struct {
	// Key is the name of the key. It is also part of the URL path when accessed
	// via the API.
	Key string

	// CreateIndex holds the index corresponding the creation of this KVPair. This
	// is a read-only field.
	CreateIndex uint64

	// ModifyIndex is used for the Check-And-Set operations and can also be fed
	// back into the WaitIndex of the QueryOptions in order to perform blocking
	// queries.
	ModifyIndex uint64

	// LockIndex holds the index corresponding to a lock on this key, if any. This
	// is a read-only field.
	LockIndex uint64

	// Flags are any user-defined flags on the key. It is up to the implementer
	// to check these values, since Consul does not treat them specially.
	Flags uint64

	// Value is the value for the key. This can be any value, but it will be
	// base64 encoded upon transport.
	Value []byte

	// Session is a string representing the ID of the session. Any other
	// interactions with this key over the same session must specify the same
	// session ID.
	Session string
}

func Init(ctx context.Context, host, prefix string, port int) context.Context {
	cfg := ConsulConfig{host, prefix, port}
	return context.WithValue(ctx, consul, cfg)
}

func getCurrentAppConfig(ctx context.Context) (value map[string]interface{}, err error) {
	consulConfig := ctx.Value(consul).(*ConsulConfig)
	currentAppConfig, err := consulConfig.getConfig()
	return currentAppConfig, err

}

func putCurrentAppConfig(ctx context.Context, value map[string]interface{}) (err error) {
	return nil
}

func (c *ConsulConfig) getConfig() (value map[string]interface{}, err error) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/v1/kv/%s", c.host, c.port, c.prefix))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 404 {
		resp.Body.Close()
		return make(map[string]interface{}, 0), nil
	} else if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}
	var entry *ConsulEntry
	err = decodeBody(resp, &entry)
	if err != nil {
		return nil, err
	}
	config := make(map[string]interface{}, 1)
	config[entry.Key] = entry.Value
	return config, err
}

func (c *ConsulConfig) putConfig(value map[string]interface{}) error {
	body, err := encodeBody(value)
	if err != nil {
		return err
	}
	_, err = http.NewRequest("PUT", fmt.Sprintf("http://%s:%d/v1/kv/%s", c.host, c.port,
		c.prefix), body)
	if err != nil {
		return err
	}
	return nil
}

func decodeBody(resp *http.Response, out interface{}) error {
	dec := json.NewDecoder(resp.Body)
	return dec.Decode(out)
}

func encodeBody(value map[string]interface{}) (io.Reader, error) {
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(value); err != nil {
		return nil, err
	}
	return buf, nil
}
