package main

import (
	"testing"
)

func TestCleanInput(t *testing.T)  {
	cases := []struct {
		input string
		expected []string
	}{
		{
			input: " hello world ",
			expected: []string{"hello", "world"},
		},
		{
			input: "",
			expected: []string{},
		},
	}

	for _, testCase := range cases {
		actual := cleanInput(testCase.input)
		
		if len(actual) != len(testCase.expected) {
			t.Errorf("lengths of actual vs expected did not match")
		}

		for i := range actual {
			word := actual[i]
			expectedWord := testCase.expected[i]
			if word != expectedWord {
				t.Errorf("word: %s did not match expectedWord: %s", word, expectedWord)
			}
		}
	}
}

