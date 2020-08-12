/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/nirmata/kyverno/pkg/api/kyverno/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ClusterKyvernoPolicyReportLister helps list ClusterKyvernoPolicyReports.
type ClusterKyvernoPolicyReportLister interface {
	// List lists all ClusterKyvernoPolicyReports in the indexer.
	List(selector labels.Selector) (ret []*v1.ClusterKyvernoPolicyReport, err error)
	// Get retrieves the ClusterKyvernoPolicyReport from the index for a given name.
	Get(name string) (*v1.ClusterKyvernoPolicyReport, error)
	ClusterKyvernoPolicyReportListerExpansion
}

// clusterKyvernoPolicyReportLister implements the ClusterKyvernoPolicyReportLister interface.
type clusterKyvernoPolicyReportLister struct {
	indexer cache.Indexer
}

// NewClusterKyvernoPolicyReportLister returns a new ClusterKyvernoPolicyReportLister.
func NewClusterKyvernoPolicyReportLister(indexer cache.Indexer) ClusterKyvernoPolicyReportLister {
	return &clusterKyvernoPolicyReportLister{indexer: indexer}
}

// List lists all ClusterKyvernoPolicyReports in the indexer.
func (s *clusterKyvernoPolicyReportLister) List(selector labels.Selector) (ret []*v1.ClusterKyvernoPolicyReport, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.ClusterKyvernoPolicyReport))
	})
	return ret, err
}

// Get retrieves the ClusterKyvernoPolicyReport from the index for a given name.
func (s *clusterKyvernoPolicyReportLister) Get(name string) (*v1.ClusterKyvernoPolicyReport, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("clusterkyvernopolicyreport"), name)
	}
	return obj.(*v1.ClusterKyvernoPolicyReport), nil
}
