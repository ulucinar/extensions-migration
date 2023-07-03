// Copyright 2023 Upbound Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package configuration has generic functions to get the new provider names
// from compositions and managed resources.
package configuration

import (
	"fmt"
	"strings"

	xpmetav1 "github.com/crossplane/crossplane/apis/pkg/meta/v1"
	xpmetav1alpha1 "github.com/crossplane/crossplane/apis/pkg/meta/v1alpha1"
	xppkgv1 "github.com/crossplane/crossplane/apis/pkg/v1"
	xppkgv1beta1 "github.com/crossplane/crossplane/apis/pkg/v1beta1"
	"github.com/pkg/errors"
	"github.com/upbound/upjet/pkg/migration"
)

const (
	awsPackage   = "xpkg.upbound.io/upbound/provider-aws"
	azurePackage = "xpkg.upbound.io/upbound/provider-azure"
	gcpPackage   = "xpkg.upbound.io/upbound/provider-gcp"
)

type mRPreProcessor struct {
	ProviderNames map[string]struct{}
}

func NewMRPreProcessor() *mRPreProcessor {
	return &mRPreProcessor{
		ProviderNames: map[string]struct{}{},
	}
}

type compositionPreProcessor struct {
	ProviderNames map[string]struct{}
}

func NewCompositionPreProcessor() *compositionPreProcessor {
	return &compositionPreProcessor{
		ProviderNames: map[string]struct{}{},
	}
}

// GetSSOPNameFromManagedResource collects the new provider name from MR
func (mp *mRPreProcessor) GetSSOPNameFromManagedResource(u migration.UnstructuredWithMetadata) error {
	for _, pn := range getProviderAndServiceName(u.Object.GroupVersionKind().Group) {
		mp.ProviderNames[pn] = struct{}{}
	}
	return nil
}

// GetSSOPNameFromComposition collects the new provider name from Composition
func (cp *compositionPreProcessor) GetSSOPNameFromComposition(u migration.UnstructuredWithMetadata) error {
	composition, err := migration.ToComposition(u.Object)
	if err != nil {
		return errors.Wrap(err, "unstructured object cannot be converted to composition")
	}
	for _, composedTemplate := range composition.Spec.Resources {
		composedUnstructured, err := migration.FromRawExtension(composedTemplate.Base)
		if err != nil {
			return errors.Wrap(err, "resource raw cannot convert to unstructured")
		}
		for _, pn := range getProviderAndServiceName(composedUnstructured.GroupVersionKind().Group) {
			cp.ProviderNames[pn] = struct{}{}
		}
	}
	return nil
}

func getProviderAndServiceName(name string) []string {
	parts := strings.Split(name, ".")
	if len(parts) > 3 {
		provider := ""
		switch parts[1] {
		case "aws":
			provider = "provider-aws"
		case "gcp":
			provider = "provider-gcp"
		case "azure":
			provider = "provider-azure"
		default:
			return nil
		}
		service := parts[0]
		return []string{fmt.Sprintf("%s-%s", provider, service), fmt.Sprintf("provider-family-%s", parts[1])}
	}
	return nil
}

type ConfigMetaParameters struct {
	FamilyVersion        string
	Monolith             string
	CompositionProcessor *compositionPreProcessor
}

type ConfigPkgParameters struct {
	PackageURL string
}

type LockParameters struct{}

func (cm *ConfigMetaParameters) ConfigurationMetadataV1(c *xpmetav1.Configuration) error {
	var convertedList []xpmetav1.Dependency

	for _, provider := range c.Spec.DependsOn {
		if *provider.Provider == fmt.Sprintf("xpkg.upbound.io/upbound/%s", cm.Monolith) {
			continue
		}
		convertedList = append(convertedList, provider)
	}

	for providerName := range cm.CompositionProcessor.ProviderNames {
		dependency := xpmetav1.Dependency{
			Provider: ptrFromString(fmt.Sprintf("xpkg.upbound.io/upbound/%s", providerName)),
			Version:  fmt.Sprintf(">=%s", cm.FamilyVersion),
		}
		convertedList = append(convertedList, dependency)
	}

	c.Spec.DependsOn = convertedList
	return nil
}

func (cm *ConfigMetaParameters) ConfigurationMetadataV1Alpha1(c *xpmetav1alpha1.Configuration) error {
	var convertedList []xpmetav1alpha1.Dependency

	for _, provider := range c.Spec.DependsOn {
		if *provider.Provider == fmt.Sprintf("xpkg.upbound.io/upbound/%s", cm.Monolith) {
			continue
		}
		convertedList = append(convertedList, provider)
	}

	for providerName := range cm.CompositionProcessor.ProviderNames {
		dependency := xpmetav1alpha1.Dependency{
			Provider: ptrFromString(fmt.Sprintf("xpkg.upbound.io/upbound/%s", providerName)),
			Version:  fmt.Sprintf(">=%s", cm.FamilyVersion),
		}
		convertedList = append(convertedList, dependency)
	}

	c.Spec.DependsOn = convertedList
	return nil
}

func (cp *ConfigPkgParameters) ConfigurationPackageV1(pkg *xppkgv1.Configuration) error {
	pkg.Spec.Package = cp.PackageURL
	return nil
}

func (l *LockParameters) PackageLockV1Beta1(lock *xppkgv1beta1.Lock) error {
	for i, lp := range lock.Packages {
		if lp.Source == awsPackage || lp.Source == azurePackage || lp.Source == gcpPackage {
			lock.Packages = append(lock.Packages[:i], lock.Packages[i+1:]...)
		}
	}
	return nil
}

func ptrFromString(s string) *string {
	return &s
}
