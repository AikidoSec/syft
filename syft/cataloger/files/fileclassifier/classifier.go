package fileclassifier

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"
	"text/template"

	"github.com/anchore/syft/syft/file"

	"github.com/anchore/syft/internal"
)

type Classifier struct {
	Class                    string
	FilepathPatterns         []*regexp.Regexp
	EvidencePatternTemplates []string
}

func DefaultClassifiers() []Classifier {
	return []Classifier{
		{
			Class: "python-binary",
			FilepathPatterns: []*regexp.Regexp{
				regexp.MustCompile(`(.*/|^)python(?P<version>[0-9]+\.[0-9]+)$`),
				regexp.MustCompile(`(.*/|^)libpython(?P<version>[0-9]+\.[0-9]+).so.*$`),
			},
			EvidencePatternTemplates: []string{
				`(?m)(?P<version>{{ .version }}\.[0-9]+[-_a-zA-Z0-9]*)`,
			},
		},
		{
			Class: "cpython-source",
			FilepathPatterns: []*regexp.Regexp{
				regexp.MustCompile(`(.*/|^)patchlevel.h$`),
			},
			EvidencePatternTemplates: []string{
				`(?m)#define\s+PY_VERSION\s+"?(?P<version>[0-9\.\-_a-zA-Z]+)"?`,
			},
		},
		{
			Class: "go-binary",
			FilepathPatterns: []*regexp.Regexp{
				regexp.MustCompile(`(.*/|^)go$`),
			},
			EvidencePatternTemplates: []string{
				`(?m)go(?P<version>[0-9]+\.[0-9]+(\.[0-9]+|beta[0-9]+|alpha[0-9]+|rc[0-9]+)?)`,
			},
		},
		{
			Class: "go-binary-hint",
			FilepathPatterns: []*regexp.Regexp{
				regexp.MustCompile(`(.*/|^)VERSION$`),
			},
			EvidencePatternTemplates: []string{
				`(?m)go(?P<version>[0-9]+\.[0-9]+(\.[0-9]+|beta[0-9]+|alpha[0-9]+|rc[0-9]+)?)`,
			},
		},
		{
			Class: "busybox-binary",
			FilepathPatterns: []*regexp.Regexp{
				regexp.MustCompile(`(.*/|^)busybox$`),
			},
			EvidencePatternTemplates: []string{
				`(?m)BusyBox\s+v(?P<version>[0-9]+\.[0-9]+\.[0-9]+)`,
			},
		},
	}
}

func (c Classifier) Classify(resolver file.Resolver, location file.Location) (*file.Classification, error) {
	doesFilepathMatch, filepathNamedGroupValues := filepathMatches(c.FilepathPatterns, location)
	if !doesFilepathMatch {
		return nil, nil
	}

	contentReader, err := resolver.FileContentsByLocation(location)
	if err != nil {
		return nil, err
	}
	defer internal.CloseAndLogError(contentReader, location.AccessPath)

	// TODO: there is room for improvement here, as this may use an excessive amount of memory. Alternate approach is to leverage a RuneReader.
	contents, err := ioutil.ReadAll(contentReader)
	if err != nil {
		return nil, err
	}

	var result *file.Classification
	for _, patternTemplate := range c.EvidencePatternTemplates {
		tmpl, err := template.New("").Parse(patternTemplate)
		if err != nil {
			return nil, fmt.Errorf("unable to parse classifier template=%q : %w", patternTemplate, err)
		}

		patternBuf := &bytes.Buffer{}
		err = tmpl.Execute(patternBuf, filepathNamedGroupValues)
		if err != nil {
			return nil, fmt.Errorf("unable to render template: %w", err)
		}

		pattern, err := regexp.Compile(patternBuf.String())
		if err != nil {
			return nil, fmt.Errorf("unable to compile rendered regex=%q: %w", patternBuf.String(), err)
		}

		if !pattern.Match(contents) {
			continue
		}

		matchMetadata := internal.MatchNamedCaptureGroups(pattern, string(contents))
		if result == nil {
			result = &file.Classification{
				Class:    c.Class,
				Metadata: matchMetadata,
			}
		} else {
			for key, value := range matchMetadata {
				result.Metadata[key] = value
			}
		}
	}
	return result, nil
}

func filepathMatches(patterns []*regexp.Regexp, location file.Location) (bool, map[string]string) {
	for _, path := range []string{location.RealPath, location.AccessPath} {
		if path == "" {
			continue
		}
		for _, pattern := range patterns {
			if pattern.MatchString(path) {
				return true, internal.MatchNamedCaptureGroups(pattern, path)
			}
		}
	}
	return false, nil
}
