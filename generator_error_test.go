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
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeneratorError(t *testing.T) {
	t.Run("GeneratorError", func(t *testing.T) {

		t.Run("Handles nil cause", func(t *testing.T) {
			err := generatorError(nil, "msg")
			require.Equal(t, "msg", err.Error())
		})

		t.Run("Formats", func(t *testing.T) {
			err := generatorError(errors.New("cause"), "msg %s", "arg")
			require.Equal(t, "msg arg\ncaused by cause", err.Error())
		})
	})
}
