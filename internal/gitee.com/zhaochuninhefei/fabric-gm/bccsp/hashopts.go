/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bccsp

import "fmt"

/*
 * bccsp/hashopts.go 提供了对`bccsp.HashOpts`的SM3实现
 */

// SM3Opts 国密 SM3.
type SM3Opts struct {
}

// Algorithm 国密 sm3 算法
func (opts *SM3Opts) Algorithm() string {
	return SM3
}

// GetHashOpt returns the HashOpts corresponding to the passed hash function
func GetHashOpt(hashFunction string) (HashOpts, error) {
	switch hashFunction {
	case SM3:
		return &SM3Opts{}, nil
	}
	return nil, fmt.Errorf("hash function not recognized [%s]", hashFunction)
}
