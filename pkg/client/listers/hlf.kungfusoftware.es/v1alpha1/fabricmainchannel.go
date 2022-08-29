// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/kfsoftware/hlf-operator/api/hlf.kungfusoftware.es/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// FabricMainChannelLister helps list FabricMainChannels.
// All objects returned here must be treated as read-only.
type FabricMainChannelLister interface {
	// List lists all FabricMainChannels in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.FabricMainChannel, err error)
	// FabricMainChannels returns an object that can list and get FabricMainChannels.
	FabricMainChannels(namespace string) FabricMainChannelNamespaceLister
	FabricMainChannelListerExpansion
}

// fabricMainChannelLister implements the FabricMainChannelLister interface.
type fabricMainChannelLister struct {
	indexer cache.Indexer
}

// NewFabricMainChannelLister returns a new FabricMainChannelLister.
func NewFabricMainChannelLister(indexer cache.Indexer) FabricMainChannelLister {
	return &fabricMainChannelLister{indexer: indexer}
}

// List lists all FabricMainChannels in the indexer.
func (s *fabricMainChannelLister) List(selector labels.Selector) (ret []*v1alpha1.FabricMainChannel, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.FabricMainChannel))
	})
	return ret, err
}

// FabricMainChannels returns an object that can list and get FabricMainChannels.
func (s *fabricMainChannelLister) FabricMainChannels(namespace string) FabricMainChannelNamespaceLister {
	return fabricMainChannelNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// FabricMainChannelNamespaceLister helps list and get FabricMainChannels.
// All objects returned here must be treated as read-only.
type FabricMainChannelNamespaceLister interface {
	// List lists all FabricMainChannels in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.FabricMainChannel, err error)
	// Get retrieves the FabricMainChannel from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.FabricMainChannel, error)
	FabricMainChannelNamespaceListerExpansion
}

// fabricMainChannelNamespaceLister implements the FabricMainChannelNamespaceLister
// interface.
type fabricMainChannelNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all FabricMainChannels in the indexer for a given namespace.
func (s fabricMainChannelNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.FabricMainChannel, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.FabricMainChannel))
	})
	return ret, err
}

// Get retrieves the FabricMainChannel from the indexer for a given namespace and name.
func (s fabricMainChannelNamespaceLister) Get(name string) (*v1alpha1.FabricMainChannel, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("fabricmainchannel"), name)
	}
	return obj.(*v1alpha1.FabricMainChannel), nil
}