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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBytesToString(t *testing.T) {
	{
		bs := []byte{0x61, 0x62}
		want := "ab"
		got := BytesToString(bs)
		assert.Equal(t, want, got)
	}

	{
		bs := []byte{}
		want := ""
		got := BytesToString(bs)
		assert.Equal(t, want, got)
	}
}

func TestSting(t *testing.T) {
	{
		want := []byte{0x61, 0x62}
		got := StringToBytes("ab")
		assert.Equal(t, want, got)
	}

	{
		want := []byte{}
		got := StringToBytes("")
		assert.Equal(t, want, got)
	}
}

func TestStingToBytes(t *testing.T) {
	{
		want := []byte{0x53, 0x45, 0x4c, 0x45, 0x43, 0x54, 0x20, 0x2a, 0x20, 0x46, 0x52, 0x4f, 0x4d, 0x20, 0x74, 0x32}
		got := StringToBytes("SELECT * FROM t2")
		assert.Equal(t, want, got)
	}
}
