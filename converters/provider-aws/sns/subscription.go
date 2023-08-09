// Copyright 2022 Upbound Inc.
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

package sns

import (
	srcv1beta1 "github.com/crossplane-contrib/provider-aws/apis/sns/v1beta1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	targetv1beta1 "github.com/upbound/provider-aws/apis/sns/v1beta1"
	"github.com/upbound/upjet/pkg/migration"
)

func SubscriptionResource(mg resource.Managed) ([]resource.Managed, error) {
	source := mg.(*srcv1beta1.Subscription)
	target := &targetv1beta1.TopicSubscription{}
	if _, err := migration.CopyInto(source, target, targetv1beta1.TopicPolicy_GroupVersionKind, "spec.forProvider.tags"); err != nil {
		return nil, errors.Wrap(err, "failed to copy source into target")
	}

	target.Spec.ForProvider.Region = &source.Spec.ForProvider.Region
	target.Spec.ForProvider.Endpoint = &source.Spec.ForProvider.Endpoint
	target.Spec.ForProvider.Protocol = &source.Spec.ForProvider.Protocol
	target.Spec.ForProvider.TopicArnRef.Name = source.Spec.ForProvider.TopicARNRef.Name

	return []resource.Managed{
		target,
	}, nil
}
