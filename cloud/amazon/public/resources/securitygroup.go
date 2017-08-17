//// Copyright © 2017 The Kubicorn Authors
////
//// Licensed under the Apache License, Version 2.0 (the "License");
//// you may not use this file except in compliance with the License.
//// You may obtain a copy of the License at
////
////     http://www.apache.org/licenses/LICENSE-2.0
////
//// Unless required by applicable law or agreed to in writing, software
//// distributed under the License is distributed on an "AS IS" BASIS,
//// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//// See the License for the specific language governing permissions and
//// limitations under the License.
//
package resources
//
//import (
//	"fmt"
//	"strconv"
//
//	"github.com/aws/aws-sdk-go/service/ec2"
//	"github.com/kris-nova/kubicorn/apis/cluster"
//	"github.com/kris-nova/kubicorn/cloud"
//	"github.com/kris-nova/kubicorn/cutil/compare"
//	"github.com/kris-nova/kubicorn/cutil/logger"
//)
//
//var _ cloud.Resource = &SecurityGroup{}
//
//type Rule struct {
//	IngressFromPort int
//	IngressToPort   int
//	IngressSource   string
//	IngressProtocol string
//}
//type SecurityGroup struct {
//	Shared
//	Firewall   *cluster.Firewall
//	ServerPool *cluster.ServerPool
//	Rules      []*Rule
//}
//
//const (
//	KubicornAutoCreatedGroup = "A fabulous security group created by Kubicorn for cluster [%s]"
//)
//
//func (r *SecurityGroup) Actual(known *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
//	logger.Debug("securitygroup.Actual")
//	actual := &SecurityGroup{
//		Shared: Shared{
//			Name:        r.Name,
//			Tags:        make(map[string]string),
//			TagResource: r.TagResource,
//		},
//	}
//
//	if r.Firewall.Identifier != "" {
//		input := &ec2.DescribeSecurityGroupsInput{
//			GroupIds: []*string{&r.Firewall.Identifier},
//		}
//		output, err := Sdk.Ec2.DescribeSecurityGroups(input)
//		if err != nil {
//			return nil, nil, err
//		}
//		lsn := len(output.SecurityGroups)
//		if lsn != 1 {
//			return nil, nil, fmt.Errorf("Found [%d] Security Groups for ID [%s]", lsn, r.Firewall.Identifier)
//		}
//		sg := output.SecurityGroups[0]
//		for _, rule := range sg.IpPermissions {
//			actual.Rules = append(actual.Rules, &Rule{
//				IngressFromPort: int(*rule.FromPort),
//				IngressToPort:   int(*rule.ToPort),
//				IngressSource:   *rule.IpRanges[0].CidrIp,
//				IngressProtocol: *rule.IpProtocol,
//			})
//		}
//		for _, tag := range sg.Tags {
//			key := *tag.Key
//			val := *tag.Value
//			actual.Tags[key] = val
//		}
//		actual.CloudID = *sg.GroupId
//		actual.Name = *sg.GroupName
//	}
//
//	r.CachedActual = actual
//	return known, actual, nil
//}
//
//func (r *SecurityGroup) Expected(known *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
//	logger.Debug("securitygroup.Expected")
//	if r.CachedExpected != nil {
//		logger.Debug("Using cached Security Group [expected]")
//		return nil, r.CachedExpected, nil
//	}
//	expected := &SecurityGroup{
//		Shared: Shared{
//			Tags: map[string]string{
//				"Name":              r.Name,
//				"KubernetesCluster": known.Name,
//			},
//			CloudID:     r.Firewall.Identifier,
//			Name:        r.Firewall.Name,
//			TagResource: r.TagResource,
//		},
//	}
//	for _, rule := range r.Firewall.IngressRules {
//		inPort, err := strToInt(rule.IngressToPort)
//		if err != nil {
//			return nil, nil, err
//		}
//		outPort, err := strToInt(rule.IngressFromPort)
//		if err != nil {
//			return nil, nil, err
//		}
//		expected.Rules = append(expected.Rules, &Rule{
//			IngressSource:   rule.IngressSource,
//			IngressToPort:   inPort,
//			IngressFromPort: outPort,
//			IngressProtocol: rule.IngressProtocol,
//		})
//		fmt.Println(rule.IngressToPort)
//	}
//
//	fmt.Println(expected.Rules)
//	fmt.Println(expected.Name)
//	r.CachedExpected = expected
//	return known, expected, nil
//}
//func (r *SecurityGroup) Apply(actual, expected cloud.Resource, applyCluster *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
//	logger.Debug("securitygroup.Apply")
//	applyResource := expected.(*SecurityGroup)
//	isEqual, err := compare.IsEqual(actual.(*SecurityGroup), expected.(*SecurityGroup))
//	if err != nil {
//		return nil, nil, err
//	}
//	if isEqual {
//		return applyCluster, applyResource, nil
//	}
//
//	input := &ec2.CreateSecurityGroupInput{
//		GroupName:   &expected.(*SecurityGroup).Name,
//		VpcId:       &applyCluster.Network.Identifier,
//		Description: S(fmt.Sprintf(KubicornAutoCreatedGroup, applyCluster.Name)),
//	}
//	output, err := Sdk.Ec2.CreateSecurityGroup(input)
//	if err != nil {
//		return nil, nil, err
//	}
//	logger.Info("Created Security Group [%s]", *output.GroupId)
//
//	newResource := &SecurityGroup{}
//	newResource.CloudID = *output.GroupId
//	newResource.Name = expected.(*SecurityGroup).Name
//
//	for _, expectedRule := range expected.(*SecurityGroup).Rules {
//		fmt.Println(expectedRule)
//		fmt.Println(expected.(*SecurityGroup).Name)
//		input := &ec2.AuthorizeSecurityGroupIngressInput{
//			GroupId:    &newResource.CloudID,
//			ToPort:     I64(expectedRule.IngressToPort),
//			FromPort:   I64(expectedRule.IngressFromPort),
//			CidrIp:     &expectedRule.IngressSource,
//			IpProtocol: S(expectedRule.IngressProtocol),
//		}
//		fmt.Println(1)
//		_, err := Sdk.Ec2.AuthorizeSecurityGroupIngress(input)
//		if err != nil {
//			return nil, nil, err
//		}
//		newResource.Rules = append(newResource.Rules, &Rule{
//			IngressSource:   expectedRule.IngressSource,
//			IngressToPort:   expectedRule.IngressToPort,
//			IngressFromPort: expectedRule.IngressFromPort,
//			IngressProtocol: expectedRule.IngressProtocol,
//		})
//
//		renderedCluster, err := r.render(newResource, applyCluster)
//		if err != nil {
//			return nil, nil, err
//		}
//		return renderedCluster, newResource, nil
//
//	}
//	return applyCluster, newResource, nil
//}
//func (r *SecurityGroup) Delete(actual cloud.Resource, known *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
//	logger.Debug("securitygroup.Delete")
//	deleteResource := actual.(*SecurityGroup)
//	if deleteResource.CloudID == "" {
//		return nil, nil, fmt.Errorf("Unable to delete Security Group resource without ID [%s]", deleteResource.Name)
//	}
//
//	input := &ec2.DeleteSecurityGroupInput{
//		GroupId: &actual.(*SecurityGroup).CloudID,
//	}
//	_, err := Sdk.Ec2.DeleteSecurityGroup(input)
//	if err != nil {
//		return nil, nil, err
//	}
//	logger.Info("Deleted Security Group [%s]", actual.(*SecurityGroup).CloudID)
//
//	newResource := &SecurityGroup{}
//	newResource.Tags = actual.(*SecurityGroup).Tags
//	newResource.Name = actual.(*SecurityGroup).Name
//
//	renderedCluster, err := r.render(newResource, known)
//	if err != nil {
//		return nil, nil, err
//	}
//	return renderedCluster, newResource, nil
//
//}
//
//func (r *SecurityGroup) render(renderResource cloud.Resource, renderCluster *cluster.Cluster) (*cluster.Cluster, error) {
//	logger.Debug("securitygroup.Render")
//	found := false
//	for i := 0; i < len(renderCluster.ServerPools); i++ {
//		for j := 0; j < len(renderCluster.ServerPools[i].Firewalls); j++ {
//			if renderCluster.ServerPools[i].Firewalls[j].Name == renderResource.(*SecurityGroup).Name {
//				found = true
//				renderCluster.ServerPools[i].Firewalls[j].Identifier = renderResource.(*SecurityGroup).CloudID
//				for _, renderRule := range renderResource.(*SecurityGroup).Rules {
//					renderCluster.ServerPools[i].Firewalls[j].IngressRules = append(renderCluster.ServerPools[i].Firewalls[j].IngressRules, &cluster.IngressRule{
//						IngressSource:   renderRule.IngressSource,
//						IngressFromPort: strconv.Itoa(renderRule.IngressFromPort),
//						IngressToPort:   strconv.Itoa(renderRule.IngressToPort),
//						IngressProtocol: renderRule.IngressProtocol,
//					})
//				}
//			}
//		}
//	}
//
//	if !found {
//		for i := 0; i < len(renderCluster.ServerPools); i++ {
//			if renderCluster.ServerPools[i].Name == r.ServerPool.Name {
//				found = true
//				var rules []*cluster.IngressRule
//				for _, renderRule := range renderResource.(*SecurityGroup).Rules {
//					rules = append(rules, &cluster.IngressRule{
//						IngressSource:   renderRule.IngressSource,
//						IngressFromPort: strconv.Itoa(renderRule.IngressFromPort),
//						IngressToPort:   strconv.Itoa(renderRule.IngressToPort),
//						IngressProtocol: renderRule.IngressProtocol,
//					})
//				}
//				renderCluster.ServerPools[i].Firewalls = append(renderCluster.ServerPools[i].Firewalls, &cluster.Firewall{
//					Name:         renderResource.(*SecurityGroup).Name,
//					Identifier:   renderResource.(*SecurityGroup).CloudID,
//					IngressRules: rules,
//				})
//
//			}
//		}
//	}
//
//	if !found {
//		var rules []*cluster.IngressRule
//		for _, renderRule := range renderResource.(*SecurityGroup).Rules {
//			rules = append(rules, &cluster.IngressRule{
//				IngressSource:   renderRule.IngressSource,
//				IngressFromPort: strconv.Itoa(renderRule.IngressFromPort),
//				IngressToPort:   strconv.Itoa(renderRule.IngressToPort),
//				IngressProtocol: renderRule.IngressProtocol,
//			})
//		}
//		firewalls := []*cluster.Firewall{
//			{
//				Name:         renderResource.(*SecurityGroup).Name,
//				Identifier:   renderResource.(*SecurityGroup).CloudID,
//				IngressRules: rules,
//			},
//		}
//		renderCluster.ServerPools = append(renderCluster.ServerPools, &cluster.ServerPool{
//			Name:       r.ServerPool.Name,
//			Identifier: r.ServerPool.Identifier,
//			Firewalls:  firewalls,
//		})
//	}
//
//	return renderCluster, nil
//}
//
//func strToInt(s string) (int, error) {
//	i, err := strconv.Atoi(s)
//	if err != nil {
//		return 0, fmt.Errorf("failed to convert string to int err: ", err)
//	}
//	return i, nil
//}
