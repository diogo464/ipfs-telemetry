package telemetry

import (
	"context"

	"github.com/diogo464/telemetry/metrics"
	"github.com/libp2p/go-libp2p/core/peer"
)

type serviceAccessControl struct {
	aclmetrics *metrics.AclMetrics
	accessType ServiceAccessType
	whitelist  map[peer.ID]struct{}
}

func newServiceAccessControl(accessType ServiceAccessType, whitelist map[peer.ID]struct{}, aclMetrics *metrics.AclMetrics) *serviceAccessControl {
	return &serviceAccessControl{
		aclmetrics: aclMetrics,
		accessType: accessType,
		whitelist:  whitelist,
	}
}

func (s *serviceAccessControl) isAllowed(id peer.ID) bool {
	switch s.accessType {
	case ServiceAccessPublic:
		s.aclmetrics.AllowedRequests.Add(context.Background(), 1)
		return true
	case ServiceAccessRestricted:
		_, ok := s.whitelist[id]
		if ok {
			s.aclmetrics.AllowedRequests.Add(context.Background(), 1)
		} else {
			s.aclmetrics.BlockedRequests.Add(context.Background(), 1)
		}
		return ok
	default:
		s.aclmetrics.BlockedRequests.Add(context.Background(), 1)
		return false
	}
}
