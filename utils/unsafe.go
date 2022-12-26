/*
Copyright 2021 RadonDB.

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

package utils

import (
	"fmt"
	"hash/fnv"
	"reflect"
	"unsafe"

	"k8s.io/apimachinery/pkg/util/rand"
)

// BytesToString casts slice to string without copy
func BytesToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}

	p := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&b)).Data)
	var s string
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&s))
	sh.Data = uintptr(p)
	sh.Cap = len(b)
	sh.Len = len(b)
	return s
}

// StringToBytes casts string to slice without copy
func StringToBytes(s string) []byte {
	if len(s) == 0 {
		return []byte{}
	}

	p := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&s)).Data)
	var b []byte
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh.Data = uintptr(p)
	sh.Cap = len(s)
	sh.Len = len(s)
	return b
}

// Caculate hash value of string
func Hash(s string) (string, error) {
	hash := fnv.New32()
	if _, err := hash.Write([]byte(s)); err != nil {
		return "", err
	}
	return rand.SafeEncodeString(fmt.Sprint(hash.Sum32())), nil
}
