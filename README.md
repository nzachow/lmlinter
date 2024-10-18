# lmlinter

Golang test linter based on language models

## Overview

This Go program is designed to parse Go test files, identify test cases, and
find the corresponding function implementations. It then generates prompts to
evaluate if the test case names are appropriate based on the function
implementations and test data.

## Prerequisites

- Go 1.16 or later
- A Go test file (e.g., example_test.go) and its corresponding implementation file (e.g., example.go).

## Usage

You can run this code directly with the following command:

``sh
   go run cmd/main.go -path=path/to/your_test_file_test.go
``

Replace path/to/your_test_file_test.go with the actual path to your Go test file.

## Output

The program will output the test cases found in the specified test file along
with the corresponding function implementations. It will generate prompts to
evaluate if the test case names are appropriate.

Example output:

TestExampleFunction:
Given this implementation: 
<function implementation>

And this test case: 
<test case data>

The name 'TestExampleFunction' is a good choice?

# TODO

For now this is just a POC to understand if this concept is viable.
In the future we need to integrate with some LLM providers to automatically get the results.
