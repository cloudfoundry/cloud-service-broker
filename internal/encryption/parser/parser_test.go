package parser_test

import (
	"strings"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/parser"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Parser", func() {
	DescribeTable(
		"correct",
		func(input string, expected interface{}) {
			output, err := parser.Parse(input)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(expected))
		},
		Entry(
			"empty",
			``,
			[]parser.PasswordEntry(nil),
		),
		Entry(
			"empty list",
			`[]`,
			[]parser.PasswordEntry(nil),
		),
		Entry(
			"one with minimum lengths",
			`[{"label":"five5","password":{"secret":"01234567890123456789"},"primary":true}]`,
			[]parser.PasswordEntry{
				{
					Label:   "five5",
					Secret:  "01234567890123456789",
					Primary: true,
				},
			},
		),
		Entry(
			"one with maximum lengths",
			`[{"label":"01234567890123456789","password":{"secret":"`+strings.Repeat("a", 1024)+`"},"primary":true}]`,
			[]parser.PasswordEntry{
				{
					Label:   "01234567890123456789",
					Secret:  strings.Repeat("a", 1024),
					Primary: true,
				},
			},
		),
		Entry(
			"one with extra fields",
			`[{"randomz":"extra","label":"barfoo","password":{"secret":"01234567890123456789"},"primary":true}]`,
			[]parser.PasswordEntry{
				{
					Label:   "barfoo",
					Secret:  "01234567890123456789",
					Primary: true,
				},
			},
		),
		Entry(
			"many",
			`[{"label":"barfoo","password":{"secret":"veryverysecretpassword"},"primary":false},{"label":"barbaz","password":{"secret":"anotherveryverysecretpassword"}},{"label":"bazquz","password":{"secret":"yetanotherveryverysecretpassword"},"primary":true}]`,
			[]parser.PasswordEntry{
				{
					Label:   "barfoo",
					Secret:  "veryverysecretpassword",
					Primary: false,
				},
				{
					Label:   "barbaz",
					Secret:  "anotherveryverysecretpassword",
					Primary: false,
				},
				{
					Label:   "bazquz",
					Secret:  "yetanotherveryverysecretpassword",
					Primary: true,
				},
			},
		),
	)

	DescribeTable(
		"errors",
		func(input string, expected interface{}) {
			output, err := parser.Parse(input)
			Expect(output).To(BeEmpty())
			Expect(err).To(MatchError(expected), err.Error())
		},
		Entry(
			"bad JSON",
			`[{"stuf"`,
			`password configuration JSON error: unexpected end of JSON input`,
		),
		Entry(
			"password length too short",
			`[{"label":"barfoo","password":{"secret":"veryverysecretpassword"},"primary":false},{"label":"barbaz","password":{"secret":"anotherveryverysecretpassword"}},{"label":"bazquz","password":{"secret":"tooshort"},"primary":true}]`,
			`password configuration error: expected value to be 20-1024 characters long, but got length 8: [2].secret.password`,
		),
		Entry(
			"password length too long",
			`[{"label":"barfoo","password":{"secret":"`+strings.Repeat("p", 2000)+`"},"primary":false},{"label":"barbaz","password":{"secret":"anotherveryverysecretpassword"}},{"label":"bazquz","password":{"secret":"yetanotherveryverysecretpassword"},"primary":true}]`,
			`password configuration error: expected value to be 20-1024 characters long, but got length 2000: [0].secret.password`,
		),
		Entry(
			"label length too short",
			`[{"label":"barfoo","password":{"secret":"veryverysecretpassword"},"primary":false},{"label":"bar","password":{"secret":"anotherveryverysecretpassword"}},{"label":"bazquz","password":{"secret":"yetanotherveryverysecretpassword"},"primary":true}]`,
			`password configuration error: expected value to be 5-20 characters long, but got length 3: [1].label`,
		),
		Entry(
			"label length too long",
			`[{"label":"012345678901234567890","password":{"secret":"veryverysecretpassword"},"primary":false},{"label":"barbaz","password":{"secret":"anotherveryverysecretpassword"}},{"label":"bazquz","password":{"secret":"yetanotherveryverysecretpassword"},"primary":true}]`,
			`password configuration error: expected value to be 5-20 characters long, but got length 21: [0].label`,
		),
		Entry(
			"label duplication",
			`[{"label":"barfoo","password":{"secret":"veryverysecretpassword"},"primary":false},{"label":"barbaz","password":{"secret":"anotherveryverysecretpassword"}},{"label":"barfoo","password":{"secret":"yetanotherveryverysecretpassword"},"primary":true}]`,
			`password configuration error: duplicated value, must be unique: barfoo: [2].label`,
		),
		Entry(
			"no primary",
			`[{"label":"barfoo","password":{"secret":"veryverysecretpassword"},"primary":false},{"label":"barbaz","password":{"secret":"anotherveryverysecretpassword"}},{"label":"bazquz","password":{"secret":"yetanotherveryverysecretpassword"},"primary":false}]`,
			`password configuration error: expected exactly one primary, got none: [].primary`,
		),
		Entry(
			"multiple primaries",
			`[{"label":"barfoo","password":{"secret":"veryverysecretpassword"},"primary":true},{"label":"barbaz","password":{"secret":"anotherveryverysecretpassword"}},{"label":"bazquz","password":{"secret":"yetanotherveryverysecretpassword"},"primary":true}]`,
			`password configuration error: expected exactly one primary, got multiple: [].primary`,
		),
	)
})
