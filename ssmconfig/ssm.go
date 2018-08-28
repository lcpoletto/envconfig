// Copyright (c) 2018 Luis Carlos Poletto. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package ssmconfig

import (
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/lcpoletto/kvconfig"
)

type ssmParamReader struct {
	path   string
	ssmAPI ssmiface.SSMAPI
	params map[string]string
	once   sync.Once
}

func (spr *ssmParamReader) lookupParam(key string) (string, bool) {
	if err := spr.loadParams(); err != nil {
		panic(err)
	}
	value, ok := spr.params[strings.ToLower(key)]
	return value, ok
}

func (spr *ssmParamReader) loadParams() error {
	var err error
	spr.once.Do(func() {
		var nextToken *string
		for {
			var out *ssm.GetParametersByPathOutput
			if out, err = spr.ssmAPI.GetParametersByPath(&ssm.GetParametersByPathInput{
				NextToken:      nextToken,
				Path:           &spr.path,
				Recursive:      aws.Bool(true),
				WithDecryption: aws.Bool(true),
			}); err != nil {
				// There was an error reading data, let's return it.
				break
			}

			for _, param := range out.Parameters {
				if param != nil {
					spr.params[strings.ToLower(*param.Name)] = *param.Value
				}
			}

			nextToken = out.NextToken
			if nextToken == nil {
				break
			}
		}

	})
	return err
}

// WithSSM builds parsing options to retrieve data from AWS SSM Parameter Store.
func WithSSM(path string, ssmAPI ssmiface.SSMAPI) kvconfig.ParseOption {
	spr := ssmParamReader{
		path:   path,
		ssmAPI: ssmAPI,
		params: make(map[string]string),
	}
	return kvconfig.ParseOption{
		KeyFormat: "%s/%s",
		LookupEnv: spr.lookupParam,
	}
}
