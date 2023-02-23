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
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"regexp/syntax"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// Each expression is generated and validated this many times.
	SampleSize = 999

	// Arbitrary limit in the standard package.
	// See https://golang.org/src/regexp/syntax/parse.go?s=18885:18935#L796
	MaxSupportedRepeatCount = 1000
)

func ExampleGenerate() {
	pattern := "[ab]{5}"
	str, _ := Generate(pattern)

	if matched, _ := regexp.MatchString(pattern, str); matched {
		fmt.Println("Matches!")
	}

	// Output:
	// Matches!
}

func ExampleNewGenerator() {
	pattern := "[ab]{5}"

	// Note that this uses a constant seed, so the generated string
	// will always be the same across different runs of the program.
	// Use a more random seed for real use (e.g. time-based).
	generator, _ := NewGenerator(pattern, &GeneratorArgs{
		RngSource: rand.NewSource(0),
	})

	str := generator.Generate()

	if matched, _ := regexp.MatchString(pattern, str); matched {
		fmt.Println("Matches!")
	}

	// Output:
	// Matches!
}

func ExampleNewGenerator_perl() {
	pattern := `\d{5}`

	generator, _ := NewGenerator(pattern, &GeneratorArgs{
		Flags: syntax.Perl,
	})

	str := generator.Generate()

	if matched, _ := regexp.MatchString("[[:digit:]]{5}", str); matched {
		fmt.Println("Matches!")
	}
	// Output:
	// Matches!
}

func ExampleCaptureGroupHandler() {
	pattern := `Hello, (?P<firstname>[A-Z][a-z]{2,10}) (?P<lastname>[A-Z][a-z]{2,10})`

	generator, _ := NewGenerator(pattern, &GeneratorArgs{
		Flags: syntax.Perl,
		CaptureGroupHandler: func(index int, name string, group *syntax.Regexp, generator Generator, args *GeneratorArgs) string {
			if name == "firstname" {
				return fmt.Sprintf("FirstName (e.g. %s)", generator.Generate())
			}
			return fmt.Sprintf("LastName (e.g. %s)", generator.Generate())
		},
	})

	// Print to stderr since we're generating random output and can't assert equality.
	fmt.Fprintln(os.Stderr, generator.Generate())

	// Needed for "go test" to run this example. (Must be a blank line before.)

	// Output:
}

func TestGeneratorArgs(t *testing.T) {
	t.Parallel()

	t.Run("initialize", func(t *testing.T) {
		t.Run("Handles empty struct", func(t *testing.T) {
			args := GeneratorArgs{}

			err := args.initialize()
			require.NoError(t, err)
		})

		t.Run("Unicode groups not supported", func(t *testing.T) {
			args := &GeneratorArgs{
				Flags: syntax.UnicodeGroups,
			}

			err := args.initialize()
			require.Error(t, err, "UnicodeGroups not supported")
		})

		t.Run("Panics if repeat bounds are invalid", func(t *testing.T) {
			args := &GeneratorArgs{
				MinUnboundedRepeatCount: 2,
				MaxUnboundedRepeatCount: 1,
			}

			require.PanicsWithValue(t, "MinUnboundedRepeatCount(2) > MaxUnboundedRepeatCount(1)", func() {
				_ = args.initialize()
			})
		})

		t.Run("Allows equal repeat bounds", func(t *testing.T) {
			args := &GeneratorArgs{
				MinUnboundedRepeatCount: 1,
				MaxUnboundedRepeatCount: 1,
			}

			err := args.initialize()
			require.NoError(t, err)
		})
	})

	t.Run("Rng", func(t *testing.T) {
		t.Run("Panics if called before initialization", func(t *testing.T) {
			require.Panics(t, func() {
				args := GeneratorArgs{}
				args.Rng()
			})
		})

		t.Run("Non-nil after initialization", func(t *testing.T) {
			args := GeneratorArgs{}
			err := args.initialize()
			require.NoError(t, err)
			require.NotNil(t, args.Rng())
		})
	})
}

func TestNewGenerator(t *testing.T) {
	t.Parallel()

	t.Run("NewGenerator", func(t *testing.T) {

		t.Run("Handles nil GeneratorArgs", func(t *testing.T) {
			generator, err := NewGenerator("", nil)
			require.NotNil(t, generator)
			require.NoError(t, err)
		})

		t.Run("Handles empty GeneratorArgs", func(t *testing.T) {
			generator, err := NewGenerator("", &GeneratorArgs{})
			require.NotNil(t, generator)
			require.NoError(t, err)
		})

		t.Run("Forwards errors from args initialization", func(t *testing.T) {
			args := &GeneratorArgs{
				Flags: syntax.UnicodeGroups,
			}

			_, err := NewGenerator("", args)
			require.NotNil(t, err)
		})
	})
}

func TestGenEmpty(t *testing.T) {
	t.Parallel()

	t.Run("Empty", func(t *testing.T) {
		args := &GeneratorArgs{
			RngSource: rand.NewSource(0),
		}
		ShouldGenerateStringMatching(t, "", "^$", args)
	})
}

func TestGenLiterals(t *testing.T) {
	t.Parallel()

	t.Run("Literals", func(t *testing.T) {
		ShouldGenerateStringsMatchingThemselves(t,
			nil,
			"a",
			"abc",
		)
	})
}

func TestGenDotNotNl(t *testing.T) {
	t.Parallel()

	t.Run("DotNotNl", func(t *testing.T) {
		ShouldGenerateStringsMatchingThemselves(t, nil, ".")

		t.Run("No newlines are generated", func(t *testing.T) {
			generator, _ := NewGenerator(".", nil)

			// Not a very strong assertion, but not sure how to do better. Exploring the entire
			// generation space (2^32) takes far too long for a unit test.
			for i := 0; i < SampleSize; i++ {
				require.NotContains(t, generator.Generate(), "\n")
			}
		})
	})
}

func TestGenStringStartEnd(t *testing.T) {
	t.Parallel()

	t.Run("String start/end", func(t *testing.T) {
		args := &GeneratorArgs{
			RngSource: rand.NewSource(0),
			Flags:     0,
		}

		ShouldGenerateStringMatching(t, `^abc$`, `^abc$`, args)
		ShouldGenerateStringMatching(t, `$abc^`, `^abc$`, args)
		ShouldGenerateStringMatching(t, `a^b$c`, `^abc$`, args)
	})
}

func TestGenQuestionMark(t *testing.T) {
	t.Parallel()

	t.Run("QuestionMark", func(t *testing.T) {
		ShouldGenerateStringsMatchingThemselves(t, nil,
			"a?",
			"(abc)?",
			"[ab]?",
			".?")
	})
}

func TestGenPlus(t *testing.T) {
	t.Parallel()

	t.Run("Plus", func(t *testing.T) {
		ShouldGenerateStringsMatchingThemselves(t, nil, "a+")
	})
}

func TestGenStar(t *testing.T) {
	t.Parallel()

	t.Run("Star", func(t *testing.T) {
		ShouldGenerateStringsMatchingThemselves(t, nil, "a*")

		t.Run("HitsDefaultMin", func(t *testing.T) {
			regexp := "a*"
			args := &GeneratorArgs{
				RngSource: rand.NewSource(0),
			}
			counts := generateLenHistogram(t, regexp, DefaultMaxUnboundedRepeatCount, args)

			require.Greater(t, counts[0], 0)
		})

		t.Run("HitsCustomMin", func(t *testing.T) {
			regexp := "a*"
			args := &GeneratorArgs{
				RngSource:               rand.NewSource(0),
				MinUnboundedRepeatCount: 200,
			}
			counts := generateLenHistogram(t, regexp, DefaultMaxUnboundedRepeatCount, args)

			require.Greater(t, counts[200], 0)
			for i := 0; i < 200; i++ {
				require.Equal(t, 0, counts[i])
			}
		})

		t.Run("HitsDefaultMax", func(t *testing.T) {
			regexp := "a*"
			args := &GeneratorArgs{
				RngSource: rand.NewSource(0),
			}
			counts := generateLenHistogram(t, regexp, DefaultMaxUnboundedRepeatCount, args)

			require.Equal(t, DefaultMaxUnboundedRepeatCount+1, len(counts))
			require.Greater(t, counts[DefaultMaxUnboundedRepeatCount], 0)
		})

		t.Run("HitsCustomMax", func(t *testing.T) {
			regexp := "a*"
			args := &GeneratorArgs{
				RngSource:               rand.NewSource(0),
				MaxUnboundedRepeatCount: 200,
			}
			counts := generateLenHistogram(t, regexp, 200, args)

			require.Equal(t, 200+1, len(counts))
			require.Greater(t, counts[200], 0)
		})
	})
}

func TestGenCharClassNotNl(t *testing.T) {
	t.Parallel()

	t.Run("CharClassNotNl", func(t *testing.T) {
		ShouldGenerateStringsMatchingThemselves(t, nil,
			"[a]",
			"[abc]",
			"[a-d]",
			"[ac]",
			"[0-9]",
			"[a-z0-9]",
		)

		t.Run("No newlines are generated", func(t *testing.T) {
			// Try to narrow down the generation space. Still not a very strong assertion.
			generator, _ := NewGenerator("[^a-zA-Z0-9]", nil)
			for i := 0; i < SampleSize; i++ {
				assert.NotEqual(t, "\n", generator.Generate())
			}
		})
	})
}

func TestGenNegativeCharClass(t *testing.T) {
	t.Parallel()

	t.Run("NegativeCharClass", func(t *testing.T) {
		ShouldGenerateStringsMatchingThemselves(t, nil, "[^a-zA-Z0-9]")
	})
}

func TestGenAlternate(t *testing.T) {
	t.Parallel()

	t.Run("Alternate", func(t *testing.T) {
		ShouldGenerateStringsMatchingThemselves(t, nil,
			"a|b",
			"abc|def|ghi",
			"[ab]|[cd]",
			"foo|bar|baz", // rewrites to foo|ba[rz]
		)
	})
}

func TestGenCapture(t *testing.T) {
	t.Parallel()

	t.Run("Capture", func(t *testing.T) {
		ShouldGenerateStringMatching(t, "(abc)", "^abc$", nil)
		ShouldGenerateStringMatching(t, "()", "^$", nil)
	})
}

func TestGenConcat(t *testing.T) {
	t.Parallel()

	t.Run("Concat", func(t *testing.T) {
		ShouldGenerateStringsMatchingThemselves(t, nil, "[ab][cd]")
	})
}

func TestGenRepeat(t *testing.T) {
	t.Parallel()

	t.Run("Repeat", func(t *testing.T) {

		t.Run("Unbounded", func(t *testing.T) {
			ShouldGenerateStringsMatchingThemselves(t, nil, `a{1,}`)

			t.Run("HitsDefaultMax", func(t *testing.T) {
				regexp := "a{0,}"
				args := &GeneratorArgs{
					RngSource: rand.NewSource(0),
				}
				counts := generateLenHistogram(t, regexp, DefaultMaxUnboundedRepeatCount, args)

				require.Equal(t, DefaultMaxUnboundedRepeatCount+1, len(counts))
				require.Greater(t, counts[DefaultMaxUnboundedRepeatCount], 0)
			})

			t.Run("HitsCustomMax", func(t *testing.T) {
				regexp := "a{0,}"
				args := &GeneratorArgs{
					RngSource:               rand.NewSource(0),
					MaxUnboundedRepeatCount: 200,
				}
				counts := generateLenHistogram(t, regexp, 200, args)

				require.Equal(t, 200+1, len(counts))
				require.Greater(t, counts[200], 0)
			})
		})

		t.Run("HitsMin", func(t *testing.T) {
			regexp := "a{0,3}"
			args := &GeneratorArgs{
				RngSource: rand.NewSource(0),
			}
			counts := generateLenHistogram(t, regexp, 3, args)

			require.Equal(t, 3+1, len(counts))
			require.Greater(t, counts[0], 0)
		})

		t.Run("HitsMax", func(t *testing.T) {
			regexp := "a{0,3}"
			args := &GeneratorArgs{
				RngSource: rand.NewSource(0),
			}
			counts := generateLenHistogram(t, regexp, 3, args)

			require.Equal(t, 3+1, len(counts))
			require.Greater(t, counts[3], 0)
		})

		t.Run("IsWithinBounds", func(t *testing.T) {
			regexp := "a{5,10}"
			args := &GeneratorArgs{
				RngSource: rand.NewSource(0),
			}
			counts := generateLenHistogram(t, regexp, 10, args)

			require.Equal(t, 11, len(counts))

			for i := 0; i < 11; i++ {
				if i < 5 {
					require.Equal(t, 0, counts[i])
				} else if i < 11 {
					require.Greater(t, counts[i], 0)
				}
			}
		})
	})
}

func TestGenCharClasses(t *testing.T) {
	t.Parallel()

	t.Run("CharClasses", func(t *testing.T) {

		t.Run("Ascii", func(t *testing.T) {
			ShouldGenerateStringsMatchingThemselves(t, nil,
				"[[:alnum:]]",
				"[[:alpha:]]",
				"[[:ascii:]]",
				"[[:blank:]]",
				"[[:cntrl:]]",
				"[[:digit:]]",
				"[[:graph:]]",
				"[[:lower:]]",
				"[[:print:]]",
				"[[:punct:]]",
				"[[:space:]]",
				"[[:upper:]]",
				"[[:word:]]",
				"[[:xdigit:]]",
				"[[:^alnum:]]",
				"[[:^alpha:]]",
				"[[:^ascii:]]",
				"[[:^blank:]]",
				"[[:^cntrl:]]",
				"[[:^digit:]]",
				"[[:^graph:]]",
				"[[:^lower:]]",
				"[[:^print:]]",
				"[[:^punct:]]",
				"[[:^space:]]",
				"[[:^upper:]]",
				"[[:^word:]]",
				"[[:^xdigit:]]",
			)
		})

		t.Run("Perl", func(t *testing.T) {
			args := &GeneratorArgs{
				Flags: syntax.Perl,
			}

			ShouldGenerateStringsMatchingThemselves(t, args,
				`\d`,
				`\s`,
				`\w`,
				`\D`,
				`\S`,
				`\W`,
			)
		})
	})
}

func TestCaptureGroupHandler(t *testing.T) {
	t.Parallel()

	t.Run("CaptureGroupHandler", func(t *testing.T) {
		callCount := 0

		gen, err := NewGenerator(`(?:foo) (bar) (?P<name>baz)`, &GeneratorArgs{
			Flags: syntax.PerlX,
			CaptureGroupHandler: func(index int, name string, group *syntax.Regexp, generator Generator, args *GeneratorArgs) string {
				callCount++

				require.Less(t, index, 2)

				if index == 0 {
					require.Equal(t, "", name)
					require.Equal(t, "bar", group.String())
					require.Equal(t, "bar", generator.Generate())
					return "one"
				}

				// Index 1
				require.Equal(t, "name", name)
				require.Equal(t, "baz", group.String())
				require.Equal(t, "baz", generator.Generate())
				return "two"
			},
		})
		require.NoError(t, err)

		require.Equal(t, "foo one two", gen.Generate())
		require.Equal(t, 2, callCount)
	})
}

func ShouldGenerateStringsMatchingThemselves(t *testing.T, args *GeneratorArgs, patterns ...string) {
	for _, pattern := range patterns {
		ShouldGenerateStringMatching(t, pattern, pattern, args)
	}
}

func ShouldGenerateStringMatching(t *testing.T, pattern, expectedPattern string, args *GeneratorArgs) {
	ShouldGenerateStringMatchingTimes(t, pattern, expectedPattern, args, SampleSize)
}

func ShouldGenerateStringMatchingTimes(t *testing.T, pattern, expectedPattern string, args *GeneratorArgs, times int) {
	generator, err := NewGenerator(pattern, args)
	require.NoError(t, err)

	for i := 0; i < times; i++ {
		result := generator.Generate()
		matched, err := regexp.MatchString(expectedPattern, result)
		require.NoError(t, err)
		if !matched {
			assert.Fail(t, "string “%s” generated from /%s/ did not match /%s/.",
				result, pattern, expectedPattern)
		}
	}
}

func generateLenHistogram(t *testing.T, regexp string, maxLen int, args *GeneratorArgs) (counts []int) {
	generator, err := NewGenerator(regexp, args)
	require.NoError(t, err)

	iterations := maxLen * 4
	if SampleSize > iterations {
		iterations = SampleSize
	}

	for i := 0; i < iterations; i++ {
		str := generator.Generate()

		// Grow the slice if necessary.
		if len(str) >= len(counts) {
			newCounts := make([]int, len(str)+1)
			copy(newCounts, counts)
			counts = newCounts
		}

		counts[len(str)]++
	}

	return
}
