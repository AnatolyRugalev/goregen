/*
Copyright 2014 Zachary Klippenstein

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

package regen

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestXorShift64(t *testing.T) {
	t.Run("Int63 should never return negative numbers.", func(t *testing.T) {

		source := xorShift64Source(1)
		for i := 0; i < SampleSize; i++ {
			val := source.Int63()

			require.GreaterOrEqual(t, val, int64(0))
		}
	})

	t.Run("Should not only return zeros", func(t *testing.T) {
		source := xorShift64Source(0)
		nonZeroCount := 0

		for i := 0; i < SampleSize; i++ {
			if source.Int63() != 0 {
				nonZeroCount++
			}
		}
		require.Greater(t, nonZeroCount, 0)
	})
}
