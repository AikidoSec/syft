package kernel

import (
	"debug/elf"
	"fmt"
	"strings"

	"github.com/anchore/syft/syft/artifact"
	"github.com/anchore/syft/syft/pkg"
	"github.com/anchore/syft/syft/pkg/cataloger/generic"
	"github.com/anchore/syft/syft/pkg/cataloger/internal/unionreader"
	"github.com/anchore/syft/syft/source"
)

type parameter struct {
	description string
	ptype       string
}
type kernelModuleMetadata struct {
	kernelVersion string
	versionMagic  string
	sourceVersion string
	version       string
	author        string
	license       string
	name          string
	description   string
	parameters    map[string]parameter
}

func (k *kernelModuleMetadata) addEntry(entry []byte) error {
	if len(entry) == 0 {
		return nil
	}
	var key, value string
	parts := strings.SplitN(string(entry), "=", 2)
	if len(parts) > 0 {
		key = parts[0]
	}
	if len(parts) > 1 {
		value = parts[1]
	}

	switch key {
	case "version":
		k.version = value
	case "license":
		k.license = value
	case "author":
		k.author = value
	case "name":
		k.name = value
	case "vermagic":
		k.versionMagic = value
		fields := strings.Fields(value)
		if len(fields) > 0 {
			k.kernelVersion = fields[0]
		}
	case "srcversion":
		k.sourceVersion = value
	case "description":
		k.description = value
	case "parm":
		parts := strings.SplitN(value, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid parm entry: %s", value)
		}
		if m, ok := k.parameters[parts[0]]; !ok {
			k.parameters[parts[0]] = parameter{description: parts[1]}
		} else {
			m.description = parts[1]
		}
	case "parmtype":
		parts := strings.SplitN(value, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid parmtype entry: %s", value)
		}
		if m, ok := k.parameters[parts[0]]; !ok {
			k.parameters[parts[0]] = parameter{ptype: parts[1]}
		} else {
			m.ptype = parts[1]
		}
	}
	return nil
}

func parseKernelModuleFile(_ source.FileResolver, _ *generic.Environment, reader source.LocationReadCloser) ([]pkg.Package, []artifact.Relationship, error) {
	unionReader, err := unionreader.GetUnionReader(reader)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get union reader for file: %w", err)
	}
	metadata, err := parseKernelModuleMetadata(unionReader)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse kernel module metadata: %w", err)
	}
	if metadata.kernelVersion == "" {
		return nil, nil, nil
	}
	p := pkg.Package{
		Name:      packageName,
		Version:   metadata.kernelVersion,
		PURL:      packageURL(packageName, metadata.kernelVersion),
		Type:      pkg.KernelPkg,
		Locations: source.NewLocationSet(reader.Location),
	}

	p.SetID()
	return []pkg.Package{p}, nil, nil
}

func parseKernelModuleMetadata(r unionreader.UnionReader) (p *kernelModuleMetadata, err error) {
	// filename:       /lib/modules/5.15.0-1031-aws/kernel/zfs/zzstd.ko
	// version:        1.4.5a
	// license:        Dual BSD/GPL
	// description:    ZSTD Compression for ZFS
	// srcversion:     F1F818A6E016499AB7F826E
	// depends:        spl
	// retpoline:      Y
	// name:           zzstd
	// vermagic:       5.15.0-1031-aws SMP mod_unload modversions
	// sig_id:         PKCS#7
	// signer:         Build time autogenerated kernel key
	// sig_key:        49:A9:55:87:90:5B:33:41:AF:C0:A7:BE:2A:71:6C:D2:CA:34:E0:AE
	// sig_hashalgo:   sha512
	//
	// OR
	//
	// filename:       /home/ubuntu/eve/rootfs/lib/modules/5.10.121-linuxkit/kernel/drivers/net/wireless/realtek/rtl8821cu/8821cu.ko
	// version:        v5.4.1_28754.20180921_COEX20180712-3232
	// author:         Realtek Semiconductor Corp.
	// description:    Realtek Wireless Lan Driver
	// license:        GPL
	// srcversion:     960CCC648A0E0369171A2C9
	// depends:        cfg80211
	// retpoline:      Y
	// name:           8821cu
	// vermagic:       5.10.121-linuxkit SMP mod_unload
	p = &kernelModuleMetadata{
		parameters: make(map[string]parameter),
	}
	f, err := elf.NewFile(r)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	modinfo := f.Section(modinfoName)
	if modinfo == nil {
		return nil, fmt.Errorf("no section %s", modinfoName)
	}
	b, err := modinfo.Data()
	if err != nil {
		return nil, fmt.Errorf("error reading secion %s: %w", modinfoName, err)
	}
	var (
		entry []byte
	)
	for _, b2 := range b {
		if b2 == 0 {
			if err := p.addEntry(entry); err != nil {
				return nil, fmt.Errorf("error parsing entry %s: %w", string(entry), err)
			}
			entry = []byte{}
			continue
		}
		entry = append(entry, b2)
	}
	if err := p.addEntry(entry); err != nil {
		return nil, fmt.Errorf("error parsing entry %s: %w", string(entry), err)
	}

	return p, nil
}
