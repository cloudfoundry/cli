package app_files_test

import (
	. "github.com/cloudfoundry/cli/cf/app_files"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF Ignore", func() {
	It("excludes files based on exact path matches", func() {
		ignore := NewCfIgnore(`the-dir/the-path`)
		Expect(ignore.FileShouldBeIgnored("the-dir/the-path")).To(BeTrue())
	})

	It("excludes the contents of directories based on exact path matches", func() {
		ignore := NewCfIgnore(`dir1/dir2`)
		Expect(ignore.FileShouldBeIgnored("dir1/dir2/the-file")).To(BeTrue())
		Expect(ignore.FileShouldBeIgnored("dir1/dir2/dir3/the-file")).To(BeTrue())
	})

	It("excludes files based on star patterns", func() {
		ignore := NewCfIgnore(`dir1/*.so`)
		Expect(ignore.FileShouldBeIgnored("dir1/file1.so")).To(BeTrue())
		Expect(ignore.FileShouldBeIgnored("dir1/file2.cc")).To(BeFalse())
	})

	It("excludes files based on double-star patterns", func() {
		ignore := NewCfIgnore(`dir1/**/*.so`)
		Expect(ignore.FileShouldBeIgnored("dir1/dir2/dir3/file1.so")).To(BeTrue())
		Expect(ignore.FileShouldBeIgnored("different-dir/dir2/file.so")).To(BeFalse())
	})

	It("allows files to be explicitly included", func() {
		ignore := NewCfIgnore(`
node_modules/*
!node_modules/common
`)

		Expect(ignore.FileShouldBeIgnored("node_modules/something-else")).To(BeTrue())
		Expect(ignore.FileShouldBeIgnored("node_modules/common")).To(BeFalse())
	})

	It("applies the patterns in order from top to bottom", func() {
		ignore := NewCfIgnore(`
stuff/*
!stuff/*.c
stuff/exclude.c`)

		Expect(ignore.FileShouldBeIgnored("stuff/something.txt")).To(BeTrue())
		Expect(ignore.FileShouldBeIgnored("stuff/exclude.c")).To(BeTrue())
		Expect(ignore.FileShouldBeIgnored("stuff/include.c")).To(BeFalse())
	})

	It("ignores certain commonly ingored files by default", func() {
		ignore := NewCfIgnore(``)
		Expect(ignore.FileShouldBeIgnored(".git/objects")).To(BeTrue())

		ignore = NewCfIgnore(`!.git`)
		Expect(ignore.FileShouldBeIgnored(".git/objects")).To(BeFalse())
	})

	Describe("files named manifest.yml", func() {
		var (
			ignore CfIgnore
		)

		BeforeEach(func() {
			ignore = NewCfIgnore("")
		})

		It("ignores manifest.yml at the top level", func() {
			Expect(ignore.FileShouldBeIgnored("manifest.yml")).To(BeTrue())
		})

		It("does not ignore nested manifest.yml files", func() {
			Expect(ignore.FileShouldBeIgnored("public/assets/manifest.yml")).To(BeFalse())
		})
	})
})
