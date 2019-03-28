package config

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/untillpro/godif"
	"github.com/untillpro/igoonce/iconfig"
	"io"
	"net/http"
	"reflect"
)

// Declare s.e.
func Declare() {
	godif.Provide(&iconfig.GetCurrentAppConfig, getCurrentAppConfig)
	godif.Provide(&iconfig.PutCurrentAppConfig, putCurrentAppConfig)
}

type consulKey int

const consul consulKey = 0

type ConsulConfig struct {
	host, prefix string
	port         uint16
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

func Init(ctx context.Context, host, prefix string, port uint16) (context.Context, error) {
	if host == "" {
		return nil, errors.New("host can't be empty")
	}
	if prefix == "" {
		return nil, errors.New("passed prefix can't be empty string")
	}
	if port == 0 {
		return nil, fmt.Errorf("passed port is invalid: %d", port)
	}
	if ctx == nil {
		return nil, errors.New("passed ctx can't be nil, pass context.TODO instead")
	}
	cfg := ConsulConfig{host, prefix, port}
	return context.WithValue(ctx, consul, cfg), nil
}

//Empty implementation
func Finit() {
	//
}

func getCurrentAppConfig(ctx context.Context, config interface{}) error {
	rv := reflect.ValueOf(config)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("%s must be a pointer", reflect.ValueOf(config))
	}
	consulConfig := ctx.Value(consul).(ConsulConfig)
	err := consulConfig.getConfig(config)
	return err
}

func putCurrentAppConfig(ctx context.Context, config interface{}) error {
	if reflect.ValueOf(config).IsNil() {
		return errors.New("testConfig1 must not be nil")
	}
	consulConfig := ctx.Value(consul).(ConsulConfig)
	err := consulConfig.putConfig(config)
	return err
}

func (c *ConsulConfig) getConfig(config interface{}) error {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/v1/kv/%s", c.host, c.port, c.prefix))
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode == 404 {
		return fmt.Errorf("no testConfig1 for %s in consul", c.prefix)
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}
	err = decodeBody(resp, config)
	if err != nil {
		return err
	}
	return nil
}

func (c *ConsulConfig) putConfig(config interface{}) error {
	body, err := encodeBody(config)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:%d/v1/kv/%s", c.host, c.port,
		c.prefix), body)
	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		panic(err)
	}
	return nil
}

func decodeBody(resp *http.Response, value interface{}) error {
	var entry []*ConsulEntry
	dec := json.NewDecoder(resp.Body)
	err := dec.Decode(&entry)
	if err != nil {
		return err
	}
	err = json.Unmarshal(entry[0].Value, &value)
	if err != nil {
		return err
	}
	return nil
}

func encodeBody(value interface{}) (io.Reader, error) {
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(value); err != nil {
		return nil, err
	}
	return buf, nil
}
