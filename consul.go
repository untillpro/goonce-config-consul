package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/untillpro/igoonce/iconfig"

	"github.com/untillpro/godif"
)

// Declare s.e.
func Declare() {
	godif.Provide(&iconfig.GetCurrentAppConfig, getCurrentAppConfig)
	godif.Provide(&iconfig.PutCurrentAppConfig, putCurrentAppConfig)
}

func main() {

	a := ConsulConfig{"127.0.0.1", "a", 8500}
	b, c := a.getConfig()
	if c != nil {
		panic(c)
	}
	fmt.Println(b)
	c = a.putConfig(map[string]string{"hz": "hz", "mb": "mb"})
	if c != nil {
		panic(c)
	}
	xz, c := a.getConfig()
	if c != nil {
		panic(c)
	}
	fmt.Println(xz)
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

func getCurrentAppConfig(ctx context.Context) (value interface{}, err error) {
	consulConfig := ctx.Value(consul).(*ConsulConfig)
	currentAppConfig, err := consulConfig.getConfig()
	return currentAppConfig, err

}

func putCurrentAppConfig(ctx context.Context) (value interface{}, err error) {
	return nil, nil
}

func (c *ConsulConfig) getConfig() (value interface{}, err error) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/v1/kv/%s/?recurse", c.host, c.port, c.prefix))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 404 {
		resp.Body.Close()
		//TODO spec error for this
		return nil, fmt.Errorf("config for prefix %s is empty", c.prefix)
	} else if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}
	var entries []*ConsulEntry
	err = decodeBody(resp, &entries)
	if err != nil {
		return nil, err
	}
	config := c.consulValuesToMap(entries)
	return config, err
}

func (c *ConsulConfig) consulValuesToMap(consulValues []*ConsulEntry) map[string]string {
	config := make(map[string]string)
	for _, entry := range consulValues {
		config[substringAfter(entry.Key, c.prefix+"/")] = string(entry.Value)
	}
	return config
}

func (c *ConsulConfig) putConfig(value interface{}) error {
	//body, err := encodeBody(value)
	//if err != nil {
	//	return err
	//}
	fasdf := value.(map[string]string)
	for k, v := range fasdf {
		resp, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:%d/v1/kv/%s/%s", c.host, c.port,
			c.prefix, k), bytes.NewBufferString(v))
		if err != nil {
			return err
		}
		fmt.Println(resp.Body)
	}
	return nil
}

func decodeBody(resp *http.Response, out interface{}) error {
	dec := json.NewDecoder(resp.Body)
	return dec.Decode(out)
}

func encodeBody(obj interface{}) (io.Reader, error) {
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(obj); err != nil {
		return nil, err
	}
	return buf, nil
}

func substringAfter(value string, a string) string {
	pos := strings.LastIndex(value, a)
	if pos <= -1 {
		return ""
	}
	adjustedPos := pos + len(a)
	if adjustedPos >= len(value) {
		return ""
	}
	return value[adjustedPos:]
}
