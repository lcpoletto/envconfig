// Copyright (c) 2018 Luis Carlos Poletto. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package consulconfig

import (
	"strings"
	"sync"

	"github.com/hashicorp/consul/api"
	"github.com/lcpoletto/kvconfig"
)

type consulKVReader struct {
	path  string
	kvAPI *api.KV
	kv    map[string]string
	once  sync.Once
}

func (ckr *consulKVReader) lookupValue(key string) (string, bool) {
	if err := ckr.loadValues(); err != nil {
		panic(err)
	}
	value, ok := ckr.kv[strings.ToLower(key)]
	return value, ok
}

func (ckr *consulKVReader) loadValues() error {
	var err error
	ckr.once.Do(func() {
		var pairs api.KVPairs
		pairs, _, err = ckr.kvAPI.List(ckr.path, nil)
		if err == nil {
			// No error retrieving data, let's move on to parsing.
			for _, pair := range pairs {
				ckr.kv[strings.ToLower(pair.Key)] = string(pair.Value)
			}
		}
	})
	return err
}

// WithConsul builds parsing options to retrieve data from Hashicorp Consul.
func WithConsul(path string, kvAPI *api.KV) kvconfig.ParseOption {
	ckr := consulKVReader{
		path:  path,
		kvAPI: kvAPI,
		kv:    make(map[string]string),
	}
	return kvconfig.ParseOption{
		KeyFormat: "%s/%s",
		LookupEnv: ckr.lookupValue,
	}
}
