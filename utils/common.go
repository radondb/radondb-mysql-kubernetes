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

import "sort"

// Min returns the smallest int64 that was passed in the arguments.
func Min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

// Max returns the largest int64 that was passed in the arguments.
func Max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

// StringInArray check whether the str is in the strArray.
func StringInArray(str string, strArray []string) bool {
	sort.Strings(strArray)
	index := sort.SearchStrings(strArray, str)
	return index < len(strArray) && strArray[index] == str
}
