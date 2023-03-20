/*
Package kernel provides a concrete Cataloger implementation for linux kernel and module files.
*/
package kernel

import (
	"github.com/anchore/syft/syft/pkg/cataloger/generic"
)

type CatalogerOpts struct {
	KernelFilenameAppends       []string
	KernelModuleFilenameAppends []string
}

var kernelFiles = []string{
	"kernel",
	"kernel-*",
	"vmlinux",
	"vmlinux-*",
	"vmlinuz",
	"vmlinuz-*",
}

var kernelModuleFiles = []string{
	"*.ko",
}

// NewKernelCataloger returns a new kernel files cataloger object.
func NewKernelCataloger(opts CatalogerOpts) *generic.Cataloger {
	var fileList []string
	for _, file := range kernelFiles {
		fileList = append(fileList, "**/"+file)
	}
	for _, file := range opts.KernelFilenameAppends {
		fileList = append(fileList, "**/"+file)
	}
	return generic.NewCataloger("linux-kernel-cataloger").
		WithParserByGlobs(parseKernelFile, fileList...)
}

// NewKernelModuleCataloger returns a new kernel module files cataloger object.
func NewKernelModuleCataloger(opts CatalogerOpts) *generic.Cataloger {
	var fileList []string
	for _, file := range kernelModuleFiles {
		fileList = append(fileList, "**/"+file)
	}
	for _, file := range opts.KernelModuleFilenameAppends {
		fileList = append(fileList, "**/"+file)
	}
	return generic.NewCataloger("linux-kernel-module-cataloger").
		WithParserByGlobs(parseKernelModuleFile, fileList...)
}
